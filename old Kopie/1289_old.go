// Copyright 2018 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fileblob provides a blob implementation that uses the filesystem.
// Use OpenBucket to construct a *blob.Bucket.
//
// Open URLs
//
// For blob.Open URLs, fileblob registers for the scheme "file"; URLs start
// with "file://".
//
// The URL's Path is used as the root directory; the URL's Host is ignored.
// If os.PathSeparator != '/', any leading '/' from the Path is dropped
// and remaining '/' characters are converted to os.PathSeparator.
// No query options are supported.
// Examples:
//  - file:///a/directory
//    -> Passes "/a/directory" to OpenBucket.
//  - file://localhost/a/directory
//    -> Also passes "/a/directory".
//  - file:///c:/foo/bar
//    -> Passes "c:\foo\bar".
//  - file://localhost/c:/foo/bar
//    -> Also passes "c:\foo\bar".
//
// Escaping
//
// Go CDK supports all UTF-8 strings; to make this work with providers lacking
// full UTF-8 support, strings must be escaped (during writes) and unescaped
// (during reads). The following escapes are performed for fileblob:
//  - Blob keys: ASCII characters 0-31 are escaped to "__0x<hex>__".
//    If os.PathSeparator != "/", it is also escaped.
//    Additionally, the "/" in "../", the trailing "/" in "//", and a trailing
//    "/" is key names are escaped in the same way.
//
// As
//
// fileblob exposes the following types for As:
//  - Error: *os.PathError
package fileblob // import "gocloud.dev/blob/fileblob"

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gocloud.dev/blob"
	"gocloud.dev/blob/driver"
	"gocloud.dev/gcerrors"
	"gocloud.dev/internal/escape"
)

const defaultPageSize = 1000

func init() {
	blob.Register("file", func(_ context.Context, u *url.URL) (driver.Bucket, error) {
		return openBucket(mungeURLPath(u.Path, os.PathSeparator), nil)
	})
}

func mungeURLPath(path string, pathSeparator uint8) string {
	if pathSeparator != '/' {
		path = strings.TrimPrefix(path, "/")
		// TODO: use filepath.FromSlash instead; and remove the pathSeparator arg
		// from this function. Test Windows behavior by opening a bucket on Windows.
		// See #1075 for why Windows is disabled.
		path = strings.Replace(path, "/", string(pathSeparator), -1)
	}
	return path
}

// Options sets options for constructing a *blob.Bucket backed by fileblob.
type Options struct{}
	// URLSigner implements signing URLs (to allow access to a resource without
	// further authorization) and verifying that a given string contains
	// a signedURL produced by the URLSigner.
	// URLSigner is only required for utilizing the SignedURL API.
	URLSigner URLSigner
}

type bucket struct {
	dir  string
	opts *Options
}

// openBucket creates a driver.Bucket that reads and writes to dir.
// dir must exist.
func openBucket(dir string, opts *Options) (driver.Bucket, error) {
	dir = filepath.Clean(dir)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}
	return &bucket{dir: dir, opts: opts}, nil
}

// OpenBucket creates a *blob.Bucket backed by the filesystem and rooted at
// dir, which must exist. See the package documentation for an example.
func OpenBucket(dir string, opts *Options) (*blob.Bucket, error) {
	drv, err := openBucket(dir, opts)
	if err != nil {
		return nil, err
	}
	return blob.NewBucket(drv), nil
}

// escapeKey does all required escaping for UTF-8 strings to work the filesystem.
func escapeKey(s string) string {
	s = escape.HexEscape(s, func(r []rune, i int) bool {
		c := r[i]
		switch {
		case c < 32:
			return true
		// We're going to replace '/' with os.PathSeparator below. In order for this
		// to be reversible, we need to escape raw os.PathSeparators.
		case os.PathSeparator != '/' && c == os.PathSeparator:
			return true
		// For "../", escape the trailing slash.
		case i > 1 && c == '/' && r[i-1] == '.' && r[i-2] == '.':
			return true
		// For "//", escape the trailing slash.
		case i > 0 && c == '/' && r[i-1] == '/':
			return true
		// Escape the trailing slash in a key.
		case c == '/' && i == len(r)-1:
			return true
		}
		return false
	})
	// Replace "/" with os.PathSeparator if needed, so that the local filesystem
	// can use subdirectories.
	if os.PathSeparator != '/' {
		s = strings.Replace(s, "/", string(os.PathSeparator), -1)
	}
	return s
}

// unescapeKey reverses escapeKey.
func unescapeKey(s string) string {
	if os.PathSeparator != '/' {
		s = strings.Replace(s, string(os.PathSeparator), "/", -1)
	}
	s = escape.HexUnescape(s)
	return s
}

var errNotImplemented = errors.New("not implemented")

func (b *bucket) ErrorCode(err error) gcerrors.ErrorCode {
	switch {
	case os.IsNotExist(err):
		return gcerrors.NotFound
	case err == errNotImplemented:
		return gcerrors.Unimplemented
	default:
		return gcerrors.Unknown
	}
}

// path returns the full path for a key
func (b *bucket) path(key string) (string, error) {
	path := filepath.Join(b.dir, escapeKey(key))
	if strings.HasSuffix(path, attrsExt) {
		return "", errAttrsExt
	}
	return path, nil
}

// forKey returns the full path, os.FileInfo, and attributes for key.
func (b *bucket) forKey(key string) (string, os.FileInfo, *xattrs, error) {
	path, err := b.path(key)
	if err != nil {
		return "", nil, nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, nil, err
	}
	xa, err := getAttrs(path)
	if err != nil {
		return "", nil, nil, err
	}
	return path, info, &xa, nil
}

// ListPaged implements driver.ListPaged.
func (b *bucket) ListPaged(ctx context.Context, opts *driver.ListOptions) (*driver.ListPage, error) {

	var pageToken string
	if len(opts.PageToken) > 0 {
		pageToken = string(opts.PageToken)
	}
	pageSize := opts.PageSize
	if pageSize == 0 {
		pageSize = defaultPageSize
	}
	// If opts.Delimiter != "", lastPrefix contains the last "directory" key we
	// added. It is used to avoid adding it again; all files in this "directory"
	// are collapsed to the single directory entry.
	var lastPrefix string

	// Do a full recursive scan of the root directory.
	var result driver.ListPage
	err := filepath.Walk(b.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Couldn't read this file/directory for some reason; just skip it.
			return nil
		}
		// Skip the self-generated attribute files.
		if strings.HasSuffix(path, attrsExt) {
			return nil
		}
		// os.Walk returns the root directory; skip it.
		if path == b.dir {
			return nil
		}
		// Strip the <b.dir> prefix from path; +1 is to include the separator.
		path = path[len(b.dir)+1:]
		// Unescape the path to get the key.
		key := unescapeKey(path)
		// Skip all directories. If opts.Delimiter is set, we'll create
		// pseudo-directories later.
		// Note that returning nil means that we'll still recurse into it;
		// we're just not adding a result for the directory itself.
		if info.IsDir() {
			key += "/"
			// Avoid recursing into subdirectories if the directory name already
			// doesn't match the prefix; any files in it are guaranteed not to match.
			if len(key) > len(opts.Prefix) && !strings.HasPrefix(key, opts.Prefix) {
				return filepath.SkipDir
			}
			// Similarly, avoid recursing into subdirectories if we're making
			// "directories" and all of the files in this subdirectory are guaranteed
			// to collapse to a "directory" that we've already added.
			if lastPrefix != "" && strings.HasPrefix(key, lastPrefix) {
				return filepath.SkipDir
			}
			return nil
		}
		// Skip files/directories that don't match the Prefix.
		if !strings.HasPrefix(key, opts.Prefix) {
			return nil
		}
		var md5 []byte
		if xa, err := getAttrs(path); err == nil {
			// Note: we only have the MD5 hash for blobs that we wrote.
			// For other blobs, md5 will remain nil.
			md5 = xa.MD5
		}
		obj := &driver.ListObject{
			Key:     key,
			ModTime: info.ModTime(),
			Size:    info.Size(),
			MD5:     md5,
		}
		// If using Delimiter, collapse "directories".
		if opts.Delimiter != "" {
			// Strip the prefix, which may contain Delimiter.
			keyWithoutPrefix := key[len(opts.Prefix):]
			// See if the key still contains Delimiter.
			// If no, it's a file and we just include it.
			// If yes, it's a file in a "sub-directory" and we want to collapse
			// all files in that "sub-directory" into a single "directory" result.
			if idx := strings.Index(keyWithoutPrefix, opts.Delimiter); idx != -1 {
				prefix := opts.Prefix + keyWithoutPrefix[0:idx+len(opts.Delimiter)]
				// We've already included this "directory"; don't add it.
				if prefix == lastPrefix {
					return nil
				}
				// Update the object to be a "directory".
				obj = &driver.ListObject{
					Key:   prefix,
					IsDir: true,
				}
				lastPrefix = prefix
			}
		}
		// If there's a pageToken, skip anything before it.
		if pageToken != "" && obj.Key <= pageToken {
			return nil
		}
		// If we've already got a full page of results, set NextPageToken and stop.
		if len(result.Objects) == pageSize {
			result.NextPageToken = []byte(result.Objects[pageSize-1].Key)
			return io.EOF
		}
		result.Objects = append(result.Objects, obj)
		return nil
	})
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &result, nil
}

// As implements driver.As.
func (b *bucket) As(i interface{}) bool { return false }

// As implements driver.ErrorAs.
func (b *bucket) ErrorAs(err error, i interface{}) bool {
	if perr, ok := err.(*os.PathError); ok {
		if p, ok := i.(**os.PathError); ok {
			*p = perr
			return true
		}
	}
	return false
}

// Attributes implements driver.Attributes.
func (b *bucket) Attributes(ctx context.Context, key string) (driver.Attributes, error) {
	_, info, xa, err := b.forKey(key)
	if err != nil {
		return driver.Attributes{}, err
	}
	return driver.Attributes{
		CacheControl:       xa.CacheControl,
		ContentDisposition: xa.ContentDisposition,
		ContentEncoding:    xa.ContentEncoding,
		ContentLanguage:    xa.ContentLanguage,
		ContentType:        xa.ContentType,
		Metadata:           xa.Metadata,
		ModTime:            info.ModTime(),
		Size:               info.Size(),
		MD5:                xa.MD5,
	}, nil
}

// NewRangeReader implements driver.NewRangeReader.
func (b *bucket) NewRangeReader(ctx context.Context, key string, offset, length int64, opts *driver.ReaderOptions) (driver.Reader, error) {
	path, info, xa, err := b.forKey(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if offset > 0 {
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return nil, err
		}
	}
	r := io.Reader(f)
	if length >= 0 {
		r = io.LimitReader(r, length)
	}
	return &reader{
		r: r,
		c: f,
		attrs: driver.ReaderAttributes{
			ContentType: xa.ContentType,
			ModTime:     info.ModTime(),
			Size:        info.Size(),
		},
	}, nil
}

type reader struct {
	r     io.Reader
	c     io.Closer
	attrs driver.ReaderAttributes
}

func (r *reader) Read(p []byte) (int, error) {
	if r.r == nil {
		return 0, io.EOF
	}
	return r.r.Read(p)
}

func (r *reader) Close() error {
	if r.c == nil {
		return nil
	}
	return r.c.Close()
}

func (r *reader) Attributes() driver.ReaderAttributes {
	return r.attrs
}

func (r *reader) As(i interface{}) bool { return false }

// NewTypedWriter implements driver.NewTypedWriter.
func (b *bucket) NewTypedWriter(ctx context.Context, key string, contentType string, opts *driver.WriterOptions) (driver.Writer, error) {
	path, err := b.path(key)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return nil, err
	}
	f, err := ioutil.TempFile(filepath.Dir(path), "fileblob")
	if err != nil {
		return nil, err
	}
	if opts.BeforeWrite != nil {
		if err := opts.BeforeWrite(func(interface{}) bool { return false }); err != nil {
			return nil, err
		}
	}
	var metadata map[string]string
	if len(opts.Metadata) > 0 {
		metadata = opts.Metadata
	}
	attrs := xattrs{
		CacheControl:       opts.CacheControl,
		ContentDisposition: opts.ContentDisposition,
		ContentEncoding:    opts.ContentEncoding,
		ContentLanguage:    opts.ContentLanguage,
		ContentType:        contentType,
		Metadata:           metadata,
	}
	w := &writer{
		ctx:        ctx,
		f:          f,
		path:       path,
		attrs:      attrs,
		contentMD5: opts.ContentMD5,
		md5hash:    md5.New(),
	}
	return w, nil
}

type writer struct {
	ctx        context.Context
	f          *os.File
	path       string
	attrs      xattrs
	contentMD5 []byte
	// We compute the MD5 hash so that we can store it with the file attributes,
	// not for verification.
	md5hash hash.Hash
}

func (w *writer) Write(p []byte) (n int, err error) {
	if _, err := w.md5hash.Write(p); err != nil {
		return 0, err
	}
	return w.f.Write(p)
}

func (w *writer) Close() error {
	err := w.f.Close()
	if err != nil {
		return err
	}
	// Always delete the temp file. On success, it will have been renamed so
	// the Remove will fail.
	defer func() {
		_ = os.Remove(w.f.Name())
	}()

	// Check if the write was cancelled.
	if err := w.ctx.Err(); err != nil {
		return err
	}

	md5sum := w.md5hash.Sum(nil)
	w.attrs.MD5 = md5sum

	// Write the attributes file.
	if err := setAttrs(w.path, w.attrs); err != nil {
		return err
	}
	// Rename the temp file to path.
	if err := os.Rename(w.f.Name(), w.path); err != nil {
		_ = os.Remove(w.path + attrsExt)
		return err
	}
	return nil
}

// Delete implements driver.Delete.
func (b *bucket) Delete(ctx context.Context, key string) error {
	path, err := b.path(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if err != nil {
		return err
	}
	if err = os.Remove(path + attrsExt); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// SignedURL implements driver.SignedURL
func (b *bucket) SignedURL(ctx context.Context, key string, opts *driver.SignedURLOptions) (string, error) {
	if b.opts.URLSigner == nil {
		return "", errors.New("to use SignedURL, you must call OpenBucket with a valid Options.URLSigner")
	}
	surl, err := b.opts.URLSigner.URLFromKey(ctx, key, opts)
	if err != nil {
		return "", err
	}
	return surl.String(), nil
}

// URLSigner defines an interface for
// creating and verifying a signed URL for objects
// in a fileblob bucket. Signed URLs are typically used for
// granting access to an otherwise-protected resource without
// requiring further authentication, and callers should take care
// to restrict the creation of signed URLs as is appropriate
// for their application.
type URLSigner interface {
	// URLFromKey defines how the bucket's object key will be turned
	// into a signed URL.
	URLFromKey(ctx context.Context, key string, opts *driver.SignedURLOptions) (*url.URL, error)

	// KeyFromURL must be able to validate a URL returned from URLFromKey.
	// KeyFromURL must only return the object if if the URL is
	// both unexpired and authentic.
	KeyFromURL(ctx context.Context, surl *url.URL) (string, bool)
}

// URLSignerHMAC uses the crypto/hmac package to create a message authentication code
// using a sha256 hash. Any URLSignerHMAC with a non-empty secretKey can validate
// a signed URL created by any other URLSignerHMAC with the identical secretKey.
type URLSignerHMAC struct {
	urlScheme string
	urlHost   string
	urlPath   string
	secretKey []byte
}

// NewURLSignerHMAC creates a URLSignerHMAC.
func NewURLSignerHMAC(urlScheme, urlHost, urlPath, secretKey string) *URLSignerHMAC {
	if secretKey == "" {
		return nil
	}
	return &URLSignerHMAC{
		urlScheme: urlScheme,
		urlHost:   urlHost,
		urlPath:   urlPath,
		secretKey: []byte(secretKey),
	}
}

// URLFromKey creates a signed URL with the object key and expiry as a query params.
func (h *URLSignerHMAC) URLFromKey(ctx context.Context, key string, opts *driver.SignedURLOptions) (*url.URL, error) {
	sURL := &url.URL{
		Host:   h.urlHost,
		Scheme: h.urlScheme,
		Path:   h.urlPath,
	}
	q := sURL.Query()
	q.Set("obj", key)
	expTime := time.Now().Add(opts.Expiry).Unix()
	q.Set("expiry", strconv.FormatInt(expTime, 10))
	sURL.RawQuery = q.Encode()

	mac := string(h.getMAC(sURL.RawQuery))
	q.Set("signature", mac)
	sURL.RawQuery = q.Encode()

	return sURL, nil
}

func (h *URLSignerHMAC) getMAC(message string) []byte {
	hsh := hmac.New(sha256.New, h.secretKey)
	hsh.Write([]byte(message))
	return hsh.Sum(nil)
}

// KeyFromURL checks expiry and signature, and returns the object key
// only if the signed URL is both authentic and unexpired.
func (h *URLSignerHMAC) KeyFromURL(ctx context.Context, sURL *url.URL) (string, bool) {
	q := sURL.Query()

	exp, err := strconv.ParseInt(q.Get("expiry"), 10, 64)
	if err != nil {
		return "", false
	}
	if time.Now().Unix() > exp {
		return "", false
	}

	sig := q.Get("signature")
	q.Del("signature")
	sURL.RawQuery = q.Encode()

	if !h.checkMAC(sURL.RawQuery, sig) {
		return "", false
	}
	return q.Get("obj"), true
}

func (h *URLSignerHMAC) checkMAC(message, mac string) bool {
	expected := h.getMAC(message)
	return hmac.Equal([]byte(mac), expected)
}
