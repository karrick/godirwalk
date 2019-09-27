package godirwalk

type scanner interface {
	Err() error
	Scan() bool
	Dirent() (*Dirent, error)
}

type dirents struct {
	dd []*Dirent
	de *Dirent
}

func (d *dirents) Err() error {
	d.dd, d.de = nil, nil
	return nil
}

func (d *dirents) Dirent() (*Dirent, error) { return d.de, nil }

func (d *dirents) Scan() bool {
	if len(d.dd) > 0 {
		d.de, d.dd = d.dd[0], d.dd[1:]
		return true
	}
	return false
}
