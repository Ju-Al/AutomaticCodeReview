	if err != nil {
		return nil, err
	}
	if ch == nil {
		return nil, nil
	}
	last := uint64(0)
	c := make(chan result)
	p := newPath(last)
	p.latest.chunk = ch
	for p.level = 1; diff>>p.level > 0; p.level++ {
	}
	go f.at(ctx, at, p, c)
	for r := range c {
		if r.chunk == nil {
			if r.level == 0 {
				return r.path.latest.chunk, nil
			}
			if r.path.level < r.level {
				continue
			}
			r.path.level = r.level - 1
		} else {
			if r.path.cancel != nil {
				close(r.path.cancel)
				r.path.cancel = nil
			}
			if r.diff == 0 {
				return r.chunk, nil
			}
			if r.path.latest.level <= r.level {
				r.path.latest = r
			}
		}
		// below applies even  if  r.path.latest==maxLevel
		if r.path.latest.level == r.path.level {
			if r.path.level == 0 {
				return r.path.latest.chunk, nil
			}
			np := newPath(r.path.latest.seq - 1)
			np.level = r.path.level
			np.latest.chunk = r.path.latest.chunk
			go f.at(ctx, at, np, c)
		}
	}
	return nil, nil
}

func (f *AsyncFinder) at(ctx context.Context, at int64, p *path, c chan result) {
	for i := p.level; i > 0; i-- {
		select {
		case <-p.cancel:
			return
		default:
		}
		go func(i int) {
			seq := p.base + 1<<i
			ch, diff, err := f.get(ctx, at, seq)
			if err != nil {
				return
			}
			select {
			case c <- result{ch, p, i, seq, diff}:
			case <-time.After(time.Minute):
			}
		}(i)
	}
}

func (f *AsyncFinder) get(ctx context.Context, at int64, seq uint64) (swarm.Chunk, int64, error) {
	u, err := f.getter.Get(ctx, &index{seq})
	if err != nil {
		if !errors.Is(err, storage.ErrNotFound) {
			return nil, 0, err
		}
		return nil, 0, nil
	}
	ts, err := feeds.UpdatedAt(u)
	if err != nil {
		return nil, 0, err
	}
	diff := at - int64(ts)
	if diff < 0 {
		return nil, 0, nil
	}
	return u, diff, nil
}

// Updater encapsulates a feeds putter to generate successive updates for epoch based feeds
// it persists the last update
type Updater struct {
	*feeds.Putter
	last uint64
}

// NewUpdater constructs a feed updater
func NewUpdater(putter storage.Putter, signer crypto.Signer, topic string) (feeds.Updater, error) {
	p, err := feeds.NewPutter(putter, signer, topic)
	if err != nil {
		return nil, err
	}
	return &Updater{Putter: p}, nil
}

// Update pushes an update to the feed through the chunk stores
func (u *Updater) Update(ctx context.Context, at int64, payload []byte) error {
	err := u.Put(ctx, &index{u.last + 1}, at, payload)
	if err != nil {
		return err
	}
	u.last++
	return nil
}

func (u *Updater) Feed() *feeds.Feed {
	return u.Putter.Feed
}
