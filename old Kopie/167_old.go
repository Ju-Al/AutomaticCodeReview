package restic

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/restic/restic/backend"
	"github.com/restic/restic/chunker"
	"github.com/restic/restic/debug"
	"github.com/restic/restic/pack"
	"github.com/restic/restic/pipe"
	"github.com/restic/restic/server"

	"github.com/juju/errors"
)

const (
	maxConcurrentBlobs = 32
	maxConcurrency     = 10
)

var archiverAbortOnAllErrors = func(str string, fi os.FileInfo, err error) error { return err }
var archiverAllowAllFiles = func(string, os.FileInfo) bool { return true }

// Archiver is used to backup a set of directories.
type Archiver struct {
	s *server.Server

	blobToken chan struct{}

	Error  func(dir string, fi os.FileInfo, err error) error
	Filter func(item string, fi os.FileInfo) bool
}

// NewArchiver returns a new archiver.
func NewArchiver(s *server.Server) *Archiver {
	arch := &Archiver{
		s:         s,
		blobToken: make(chan struct{}, maxConcurrentBlobs),
	}

	for i := 0; i < maxConcurrentBlobs; i++ {
		arch.blobToken <- struct{}{}
	}

	arch.Error = archiverAbortOnAllErrors
	arch.Filter = archiverAllowAllFiles

	return arch
}

// Save stores a blob read from rd in the server.
func (arch *Archiver) Save(t pack.BlobType, id backend.ID, length uint, rd io.Reader) error {
	debug.Log("Archiver.Save", "Save(%v, %v)\n", t, id.Str())

	// test if this blob is already known
	if arch.s.Index().Has(id) {
		debug.Log("Archiver.Save", "(%v, %v) already saved\n", t, id.Str())
		return nil
	}

	// otherwise save blob
	err := arch.s.SaveFrom(t, id, length, rd)
	if err != nil {
		debug.Log("Archiver.Save", "Save(%v, %v): error %v\n", t, id.Str(), err)
		return err
	}

	debug.Log("Archiver.Save", "Save(%v, %v): new blob\n", t, id.Str())
	return nil
}

// SaveTreeJSON stores a tree in the server.
func (arch *Archiver) SaveTreeJSON(item interface{}) (backend.ID, error) {
	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')

	// check if tree has been saved before
	id := backend.Hash(data)
	if arch.s.Index().Has(id) {
		return id, nil
	}

	return arch.s.SaveJSON(pack.Tree, item)
}

func (arch *Archiver) reloadFileIfChanged(node *Node, file *os.File) (*Node, error) {
	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fi.ModTime() == node.ModTime {
		return node, nil
	}

	err = arch.Error(node.path, fi, errors.New("file has changed"))
	if err != nil {
		return nil, err
	}

	node, err = NodeFromFileInfo(node.path, fi)
	if err != nil {
		debug.Log("Archiver.SaveFile", "NodeFromFileInfo returned error for %v: %v", node.path, err)
		return nil, err
	}

	return node, nil
}

type saveResult struct {
	id    backend.ID
	bytes uint64
}

func (arch *Archiver) saveChunk(chunk *chunker.Chunk, p *Progress, token struct{}, file *os.File, resultChannel chan<- saveResult) {
	err := arch.Save(pack.Data, chunk.Digest, chunk.Length, chunk.Reader(file))
	// TODO handle error
	if err != nil {
		panic(err)
	}

	p.Report(Stat{Bytes: uint64(chunk.Length)})
	arch.blobToken <- token
	resultChannel <- saveResult{id: backend.ID(chunk.Digest), bytes: uint64(chunk.Length)}
}

func waitForResults(resultChannels [](<-chan saveResult)) ([]saveResult, error) {
	results := []saveResult{}

	for _, ch := range resultChannels {
		results = append(results, <-ch)
	}

	if len(results) != len(resultChannels) {
		return nil, fmt.Errorf("chunker returned %v chunks, but only %v blobs saved", len(resultChannels), len(results))
	}

	return results, nil
}

func updateNodeContent(node *Node, results []saveResult) error {
	debug.Log("Archiver.Save", "checking size for file %s", node.path)

	var bytes uint64
	node.Content = make([]backend.ID, len(results))

	for i, b := range results {
		node.Content[i] = b.id
		bytes += b.bytes

		debug.Log("Archiver.Save", "  adding blob %s, %d bytes", b.id.Str(), b.bytes)
	}

	if bytes != node.Size {
		return fmt.Errorf("errors saving node %q: saved %d bytes, wanted %d bytes", node.path, bytes, node.Size)
	}

	debug.Log("Archiver.SaveFile", "SaveFile(%q): %v blobs\n", node.path, len(results))

	return nil
}

var chunkerBufPool = sync.Pool{
	New: func() interface{} { return make([]byte, chunkerBufSize) },
}

// SaveFile stores the content of the file on the backend as a Blob by calling
// Save for each chunk.
func (arch *Archiver) SaveFile(p *Progress, node *Node) error {
	file, err := node.OpenForReading()
	defer file.Close()
	if err != nil {
		return err
	}

	node, err = arch.reloadFileIfChanged(node, file)
	if err != nil {
		return err
	}

	buf := chunkerBufPool.Get().([]byte)
	chnker := chunker.New(file, arch.s.Config.ChunkerPolynomial, buf, sha256.New())
	defer chunkerBufPool.Put(buf)
	resultChannels := [](<-chan saveResult){}

	for {
		chunk, err := chnker.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return errors.Annotate(err, "SaveFile() chunker.Next()")
		}

		resCh := make(chan saveResult, 1)
		go arch.saveChunk(chunk, p, <-arch.blobToken, file, resCh)
		resultChannels = append(resultChannels, resCh)
	}

	results, err := waitForResults(resultChannels)
	if err != nil {
		return err
	}

	err = updateNodeContent(node, results)
	return err
}

func (arch *Archiver) fileWorker(wg *sync.WaitGroup, p *Progress, done <-chan struct{}, entCh <-chan pipe.Entry) {
	defer func() {
		debug.Log("Archiver.fileWorker", "done")
		wg.Done()
	}()
	for {
		select {
		case e, ok := <-entCh:
			if !ok {
				// channel is closed
				return
			}

			debug.Log("Archiver.fileWorker", "got job %v", e)

			// check for errors
			if e.Error() != nil {
				debug.Log("Archiver.fileWorker", "job %v has errors: %v", e.Path(), e.Error())
				// TODO: integrate error reporting
				fmt.Fprintf(os.Stderr, "error for %v: %v\n", e.Path(), e.Error())
				// ignore this file
				e.Result() <- nil
				p.Report(Stat{Errors: 1})
				continue
			}

			node, err := NodeFromFileInfo(e.Fullpath(), e.Info())
			if err != nil {
				// TODO: integrate error reporting
				debug.Log("Archiver.fileWorker", "NodeFromFileInfo returned error for %v: %v", node.path, err)
				e.Result() <- nil
				p.Report(Stat{Errors: 1})
				continue
			}

			// try to use old node, if present
			if e.Node != nil {
				debug.Log("Archiver.fileWorker", "   %v use old data", e.Path())

				oldNode := e.Node.(*Node)
				// check if all content is still available in the repository
				contentMissing := false
				for _, blob := range oldNode.blobs {
					if ok, err := arch.s.Test(backend.Data, blob.Storage.String()); !ok || err != nil {
						debug.Log("Archiver.fileWorker", "   %v not using old data, %v (%v) is missing", e.Path(), blob.ID.Str(), blob.Storage.Str())
						contentMissing = true
						break
					}
				}

				if !contentMissing {
					node.Content = oldNode.Content
					node.blobs = oldNode.blobs
					debug.Log("Archiver.fileWorker", "   %v content is complete", e.Path())
				}
			} else {
				debug.Log("Archiver.fileWorker", "   %v no old data", e.Path())
			}

			// otherwise read file normally
			if node.Type == "file" && len(node.Content) == 0 {
				debug.Log("Archiver.fileWorker", "   read and save %v, content: %v", e.Path(), node.Content)
				err = arch.SaveFile(p, node)
				if err != nil {
					// TODO: integrate error reporting
					fmt.Fprintf(os.Stderr, "error for %v: %v\n", node.path, err)
					// ignore this file
					e.Result() <- nil
					p.Report(Stat{Errors: 1})
					continue
				}
			} else {
				// report old data size
				p.Report(Stat{Bytes: node.Size})
			}

			debug.Log("Archiver.fileWorker", "   processed %v, %d/%d blobs", e.Path(), len(node.Content), len(node.blobs))
			e.Result() <- node
			p.Report(Stat{Files: 1})
		case <-done:
			// pipeline was cancelled
			return
		}
	}
}

func (arch *Archiver) dirWorker(wg *sync.WaitGroup, p *Progress, done <-chan struct{}, dirCh <-chan pipe.Dir) {
	defer func() {
		debug.Log("Archiver.dirWorker", "done")
		wg.Done()
	}()
	for {
		select {
		case dir, ok := <-dirCh:
			if !ok {
				// channel is closed
				return
			}
			debug.Log("Archiver.dirWorker", "save dir %v\n", dir.Path())

			tree := NewTree()

			// wait for all content
			for _, ch := range dir.Entries {
				res := <-ch

				// if we get a nil pointer here, an error has happened while
				// processing this entry. Ignore it for now.
				if res == nil {
					continue
				}

				// else insert node
				node := res.(*Node)
				tree.Insert(node)

				if node.Type == "dir" {
					debug.Log("Archiver.dirWorker", "got tree node for %s: %v", node.path, node.blobs)
				}
			}

			var (
				node *Node
				err  error
			)
			if dir.Path() == "" {
				// if this is the top-level dir, only create a stub node
				node = &Node{}
			} else {
				// else create node from path and fi
				node, err = NodeFromFileInfo(dir.Path(), dir.Info())
				if err != nil {
					node.Error = err.Error()
					dir.Result() <- node
					continue
				}
			}

			id, err := arch.SaveTreeJSON(tree)
			if err != nil {
				panic(err)
			}
			debug.Log("Archiver.dirWorker", "save tree for %s: %v", dir.Path(), id.Str())

			node.Subtree = id

			dir.Result() <- node
			if dir.Path() != "" {
				p.Report(Stat{Dirs: 1})
			}
		case <-done:
			// pipeline was cancelled
			return
		}
	}
}

type archivePipe struct {
	Old <-chan WalkTreeJob
	New <-chan pipe.Job
}

func copyJobs(done <-chan struct{}, in <-chan pipe.Job, out chan<- pipe.Job) {
	var (
		// disable sending on the outCh until we received a job
		outCh chan<- pipe.Job
		// enable receiving from in
		inCh = in
		job  pipe.Job
		ok   bool
	)

	for {
		select {
		case <-done:
			return
		case job, ok = <-inCh:
			if !ok {
				// input channel closed, we're done
				debug.Log("copyJobs", "input channel closed, we're done")
				return
			}
			inCh = nil
			outCh = out
		case outCh <- job:
			outCh = nil
			inCh = in
		}
	}
}

type archiveJob struct {
	hasOld bool
	old    WalkTreeJob
	new    pipe.Job
}

func (a *archivePipe) compare(done <-chan struct{}, out chan<- pipe.Job) {
	defer func() {
		close(out)
		debug.Log("ArchivePipe.compare", "done")
	}()

	debug.Log("ArchivePipe.compare", "start")
	var (
		loadOld, loadNew bool = true, true
		ok               bool
		oldJob           WalkTreeJob
		newJob           pipe.Job
	)

	for {
		if loadOld {
			oldJob, ok = <-a.Old
			// if the old channel is closed, just pass through the new jobs
			if !ok {
				debug.Log("ArchivePipe.compare", "old channel is closed, copy from new channel")

				// handle remaining newJob
				if !loadNew {
					out <- archiveJob{new: newJob}.Copy()
				}

				copyJobs(done, a.New, out)
				return
			}

			loadOld = false
		}

		if loadNew {
			newJob, ok = <-a.New
			// if the new channel is closed, there are no more files in the current snapshot, return
			if !ok {
				debug.Log("ArchivePipe.compare", "new channel is closed, we're done")
				return
			}

			loadNew = false
		}

		debug.Log("ArchivePipe.compare", "old job: %v", oldJob.Path)
		debug.Log("ArchivePipe.compare", "new job: %v", newJob.Path())

		// at this point we have received an old job as well as a new job, compare paths
		file1 := oldJob.Path
		file2 := newJob.Path()

		dir1 := filepath.Dir(file1)
		dir2 := filepath.Dir(file2)

		if file1 == file2 {
			debug.Log("ArchivePipe.compare", "    same filename %q", file1)

			// send job
			out <- archiveJob{hasOld: true, old: oldJob, new: newJob}.Copy()
			loadOld = true
			loadNew = true
			continue
		} else if dir1 < dir2 {
			debug.Log("ArchivePipe.compare", "    %q < %q, file %q added", dir1, dir2, file2)
			// file is new, send new job and load new
			loadNew = true
			out <- archiveJob{new: newJob}.Copy()
			continue
		} else if dir1 == dir2 {
			if file1 < file2 {
				debug.Log("ArchivePipe.compare", "    %q < %q, file %q removed", file1, file2, file1)
				// file has been removed, load new old
				loadOld = true
				continue
			} else {
				debug.Log("ArchivePipe.compare", "    %q > %q, file %q added", file1, file2, file2)
				// file is new, send new job and load new
				loadNew = true
				out <- archiveJob{new: newJob}.Copy()
				continue
			}
		}

		debug.Log("ArchivePipe.compare", "    %q > %q, file %q removed", file1, file2, file1)
		// file has been removed, throw away old job and load new
		loadOld = true
	}
}

func (j archiveJob) Copy() pipe.Job {
	if !j.hasOld {
		return j.new
	}

	// handle files
	if isFile(j.new.Info()) {
		debug.Log("archiveJob.Copy", "   job %v is file", j.new.Path())

		// if type has changed, return new job directly
		if j.old.Node == nil {
			return j.new
		}

		// if file is newer, return the new job
		if j.old.Node.isNewer(j.new.Fullpath(), j.new.Info()) {
			debug.Log("archiveJob.Copy", "   job %v is newer", j.new.Path())
			return j.new
		}

		debug.Log("archiveJob.Copy", "   job %v add old data", j.new.Path())
		// otherwise annotate job with old data
		e := j.new.(pipe.Entry)
		e.Node = j.old.Node
		return e
	}

	// dirs and other types are just returned
	return j.new
}

// Snapshot creates a snapshot of the given paths. If parentID is set, this is
// used to compare the files to the ones archived at the time this snapshot was
// taken.
func (arch *Archiver) Snapshot(p *Progress, paths []string, parentID backend.ID) (*Snapshot, backend.ID, error) {
	debug.Log("Archiver.Snapshot", "start for %v", paths)

	debug.Break("Archiver.Snapshot")
	sort.Strings(paths)

	// signal the whole pipeline to stop
	done := make(chan struct{})
	var err error

	p.Start()
	defer p.Done()

	// create new snapshot
	sn, err := NewSnapshot(paths)
	if err != nil {
		return nil, nil, err
	}

	jobs := archivePipe{}

	// use parent snapshot (if some was given)
	if parentID != nil {
		sn.Parent = parentID

		// load parent snapshot
		parent, err := LoadSnapshot(arch.s, parentID)
		if err != nil {
			return nil, nil, err
		}

		// start walker on old tree
		ch := make(chan WalkTreeJob)
		go WalkTree(arch.s, parent.Tree, done, ch)
		jobs.Old = ch
	} else {
		// use closed channel
		ch := make(chan WalkTreeJob)
		close(ch)
		jobs.Old = ch
	}

	// start walker
	pipeCh := make(chan pipe.Job)
	resCh := make(chan pipe.Result, 1)
	go func() {
		err := pipe.Walk(paths, done, pipeCh, resCh)
		if err != nil {
			debug.Log("Archiver.Snapshot", "pipe.Walk returned error %v", err)
			return
		}
		debug.Log("Archiver.Snapshot", "pipe.Walk done")
	}()
	jobs.New = pipeCh

	ch := make(chan pipe.Job)
	go jobs.compare(done, ch)

	var wg sync.WaitGroup
	entCh := make(chan pipe.Entry)
	dirCh := make(chan pipe.Dir)

	// split
	wg.Add(1)
	go func() {
		pipe.Split(ch, dirCh, entCh)
		debug.Log("Archiver.Snapshot", "split done")
		close(dirCh)
		close(entCh)
		wg.Done()
	}()

	// run workers
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(2)
		go arch.fileWorker(&wg, p, done, entCh)
		go arch.dirWorker(&wg, p, done, dirCh)
	}

	// wait for all workers to terminate
	debug.Log("Archiver.Snapshot", "wait for workers")
	wg.Wait()

	debug.Log("Archiver.Snapshot", "workers terminated")

	// receive the top-level tree
	root := (<-resCh).(*Node)
	debug.Log("Archiver.Snapshot", "root node received: %v", root.Subtree.Str())
	sn.Tree = root.Subtree

	// save snapshot
	id, err := arch.s.SaveJSONUnpacked(backend.Snapshot, sn)
	if err != nil {
		return nil, nil, err
	}

	// store ID in snapshot struct
	sn.id = id
	debug.Log("Archiver.Snapshot", "saved snapshot %v", id.Str())

	// flush server
	err = arch.s.Flush()
	if err != nil {
		return nil, nil, err
	}

	// save index
	indexID, err := arch.s.SaveIndex()
	if err != nil {
		debug.Log("Archiver.Snapshot", "error saving index: %v", err)
		return nil, nil, err
	}

	debug.Log("Archiver.Snapshot", "saved index %v", indexID.Str())

	return sn, id, nil
}

func isFile(fi os.FileInfo) bool {
	if fi == nil {
		return false
	}

	return fi.Mode()&(os.ModeType|os.ModeCharDevice) == 0
}

// Scan traverses the dirs to collect Stat information while emitting progress
// information with p.
func Scan(dirs []string, p *Progress) (Stat, error) {
	p.Start()
	defer p.Done()

	var stat Stat

	for _, dir := range dirs {
		debug.Log("Scan", "Start for %v", dir)
		err := filepath.Walk(dir, func(str string, fi os.FileInfo, err error) error {
			debug.Log("Scan.Walk", "%v, fi: %v, err: %v", str, fi, err)
			// TODO: integrate error reporting
			if err != nil {
				fmt.Fprintf(os.Stderr, "error for %v: %v\n", str, err)
				return nil
			}
			if fi == nil {
				fmt.Fprintf(os.Stderr, "error for %v: FileInfo is nil
", str)
				return nil
			}
			s := Stat{}
			if isFile(fi) {
				s.Files++
				s.Bytes += uint64(fi.Size())
			} else if fi.IsDir() {
				s.Dirs++
			}

			p.Report(s)
			stat.Add(s)

			// TODO: handle error?
			return nil
		})

		debug.Log("Scan", "Done for %v, err: %v", dir, err)
		if err != nil {
			return Stat{}, err
		}
	}

	return stat, nil
}
