package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/google/uuid"
)

type Storer interface {
	Populate(entrymap map[string]*Entry, sourcemap map[string]*Source) (err error)
	AddEntry(entry *Entry) (err error)
	DeleteEntry(entry *Entry) (err error)
	AddSource(source *Source) (err error)
	DeleteSource(source *Source) (err error)
	Commit(message string) (err error)
}

// implementation
type BillyStorer struct {
	fsys    billy.Filesystem
	hashmap map[string]map[string]plumbing.Hash
	rep     *git.Repository
	bw      *bufio.Writer
	buf     *bytes.Buffer
}

func NewBillyStorer(gitdir string) *BillyStorer {
	s := &BillyStorer{
		fsys:    memfs.New(),
		hashmap: make(map[string]map[string]plumbing.Hash),
		bw:      bufio.NewWriter(nil),
		buf:     bytes.NewBuffer(nil),
	}
	if e := s.fsys.MkdirAll("entry", os.ModePerm); e != nil {
		panic(e)
	} else if e := s.fsys.MkdirAll("source", os.ModePerm); e != nil {
		panic(e)
	} else if s.hashmap["entry"] = make(map[string]plumbing.Hash); false {
		//
	} else if s.hashmap["source"] = make(map[string]plumbing.Hash); false {
		//
	} else if rep, e := s.gitStartUp(gitdir); e != nil {
		panic(e)
	} else {
		s.rep = rep
	}

	return s
}

func (s *BillyStorer) Populate(entrymap map[string]*Entry, sourcemap map[string]*Source) (err error) {
	if ref, e := s.rep.Reference(plumbing.Master, true); e != nil {
		err = e
	} else if commit_obj, e := s.rep.CommitObject(ref.Hash()); e != nil {
		err = e
	} else if tree, e := commit_obj.Tree(); e != nil {
		err = e
	} else if iter := tree.Files(); false {
		//
	} else if e := iter.ForEach(func(obj *object.File) (err error) {
		// load the sources
		if path.Dir(obj.Name) != "source" {
			return nil
		} else if objr, e := obj.Reader(); e != nil {
			err = e
		} else if source := new(Source); false {
			//
		} else if e := xml.NewDecoder(objr).Decode(source); e != nil {
			err = e
		} else if source.Id == nil || source.Updated == nil {
			err = fmt.Errorf("nil pointer dereference")
		} else if u, e := uuid.Parse(source.Id.Target); e != nil {
			err = fmt.Errorf("cannot parse source id as urn:uuid")
		} else if feed_URL := "/feed/" + u.String() + ".atom"; true {
			sourcemap[feed_URL] = source
		}
		return
	}); e != nil {
		err = e
	} else if iter = tree.Files(); false {
		//
	} else if e := iter.ForEach(func(obj *object.File) (err error) {
		if path.Dir(obj.Name) != "entry" {
			return nil
		} else if objr, e := obj.Reader(); e != nil {
			err = e
		} else if entry := new(Entry); false {
			//
		} else if e := xml.NewDecoder(objr).Decode(entry); e != nil {
			err = e
		} else if entry.Content.Body = bytes.Map(func(r rune) rune {
			if r == '\n' {
				return -1
			} else {
				return r
			}
		}, entry.Content.Body); false {
			//
		} else if _, e := entry.Validate(nil); e != nil {
			err = e
		} else if entry.Source == nil || entry.Source.Id == nil {
			err = fmt.Errorf("nil pointer dereference")
		} else if u, e := uuid.Parse(entry.Source.Id.Target); e != nil {
			err = fmt.Errorf("cannot parse source id as urn:uuid")
		} else if feed_URL := "/feed/" + u.String() + ".atom"; false {
			//
		} else if source, ok := sourcemap[feed_URL]; !ok {
			err = fmt.Errorf("entry without a source: %s", obj.Name)
		} else if entry.Source = source; false {
			// the previous source which was unmarshalled into
			// will be discarded by the garbage collector
		} else if u, e := uuid.Parse(entry.Id.Target); e != nil {
			err = fmt.Errorf("cannot parse entry id as urn:uuid")
		} else if entry_URL := "/entry/" + u.String() + ".atom"; true {
			entrymap[entry_URL] = entry
		}
		return
	}); e != nil {
		err = e
	}

	return
}

func (s *BillyStorer) AddEntry(entry *Entry) (err error) {
	if entry == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else if entry_uuid, e := uuid.Parse(entry.Id.Target); e != nil {
		err = fmt.Errorf("invalid entry id: %w", e)
	} else if f, e := s.fsys.Create(path.Join("entry", entry_uuid.String()+".atom")); e != nil {
		err = e
	} else if s.buf.Reset(); false {
		//
	} else if _, e := s.buf.WriteString(xml.Header); e != nil {
		err = e
	} else if s.bw.Reset(s.buf); false {
		//
	} else if e := entry.MarshalTo(s.bw, nil); e != nil {
		err = e
	} else if e := s.bw.Flush(); e != nil {
		err = e
	} else {
		// separating into lines makes cleaner git logs
		for {
			if l, e := s.buf.ReadBytes('>'); e != nil {
				err = e
			} else if _, e := f.Write(l); e != nil {
				err = e
			} else if b, e := s.buf.ReadByte(); e != nil {
				// probably gets tripped here
				err = e
			} else if b != '<' {
				if _, e := f.Write([]byte{b}); e != nil {
					err = e
				}
			} else {
				if _, e := f.Write([]byte{'\n', b}); e != nil {
					err = e
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return
			}
		}
		err = f.Close()
	}
	return
}

func (s *BillyStorer) DeleteEntry(entry *Entry) (err error) {
	if entry == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else if entry_uuid, e := uuid.Parse(entry.Id.Target); e != nil {
		err = fmt.Errorf("invalid entry id: %w", e)
	} else {
		delete(s.hashmap["entry"], entry_uuid.String()+".atom")
	}
	return
}

func (s *BillyStorer) AddSource(source *Source) (err error) {
	if source == nil || source.Id == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else if source_uuid, e := uuid.Parse(source.Id.Target); e != nil {
		err = fmt.Errorf("invalid entry id: %w", e)
	} else if f, e := s.fsys.Create(path.Join("source", source_uuid.String()+".atom")); e != nil {
		err = e
	} else if s.buf.Reset(); false {
		//
	} else if _, e := s.buf.WriteString(xml.Header); e != nil {
		err = e
	} else if s.bw.Reset(s.buf); false {
		//
	} else if e := source.MarshalTo(s.bw, nil, nil); e != nil {
		err = e
	} else if e := s.bw.Flush(); e != nil {
		err = e
	} else {
		for {
			if l, e := s.buf.ReadBytes('>'); e != nil {
				err = e
			} else if _, e := f.Write(l); e != nil {
				err = e
			} else if b, e := s.buf.ReadByte(); e != nil {
				// probably gets tripped here
				err = e
			} else if b != '<' {
				if _, e := f.Write([]byte{b}); e != nil {
					err = e
				}
			} else {
				if _, e := f.Write([]byte{'\n', b}); e != nil {
					err = e
				}
			}
			if errors.Is(err, io.EOF) {
				break
			} else if err != nil {
				return
			}
		}
		err = f.Close()
	}
	return nil
}

func (s *BillyStorer) DeleteSource(source *Source) (err error) {
	if source == nil {
		err = fmt.Errorf("nil pointer dereference")
	} else if source_uuid, e := uuid.Parse(source.Id.Target); e != nil {
		err = fmt.Errorf("invalid entry id: %w", e)
	} else {
		delete(s.hashmap["source"], source_uuid.String()+".atom")
	}
	return
}

func (s *BillyStorer) Commit(message string) (err error) {

	// store the objects
	if e := s.storeObjs(); e != nil {
		err = e
	} else if h, e := s.nextTree(); e != nil {
		err = e
	} else if e := s.nextCommit(h, message); e != nil {
		err = e
	}

	return
}

const pre_receive_hook = `#!/bin/sh
echo "atompub-server git backend is read-only."
exit 1`

// git
func (s *BillyStorer) gitStartUp(dir string) (rep *git.Repository, err error) {
	if r, e := git.PlainOpen(dir); e == nil {
		rep = r
	} else if r, e := git.PlainInit(dir, true); e != nil {
		err = e
	} else if e := os.Mkdir(path.Join(dir, "hooks"), os.ModePerm); e != nil {
		err = e
	} else if f, e := os.OpenFile(path.Join(dir, "hooks", "pre-receive"), os.O_WRONLY|os.O_CREATE, 0744); e != nil {
		err = e
	} else if _, e := f.WriteString(pre_receive_hook); e != nil {
		err = e
	} else if e := f.Close(); e != nil {
		err = e
	} else {
		rep = r
	}
	if err != nil {
		return
	}

	var master *plumbing.Reference
	if ref, e := rep.Reference(plumbing.Master, true); e == nil {
		master = ref
	} else if h, e := initCommit(rep); e != nil {
		err = e
	} else if ref := plumbing.NewHashReference(plumbing.Master, h); false {
		//
	} else if e := rep.Storer.SetReference(ref); e != nil {
		err = e
	} else if e := rep.Storer.SetReference(plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.Master)); e != nil {
		err = e
	} else {
		master = ref
	}
	if err != nil {
		return
	}

	if commit_obj, e := rep.CommitObject(master.Hash()); e != nil {
		err = e
	} else if tree, e := commit_obj.Tree(); e != nil {
		err = e
	} else if iter := tree.Files(); false {
		//
	} else {
		iter.ForEach(func(obj *object.File) error {
			d, f := path.Split(obj.Name)
			s.hashmap[path.Clean(d)][f] = obj.Hash
			return nil
		})
	}

	return
}

func initCommit(rep *git.Repository) (hash plumbing.Hash, err error) {
	if tree_obj := rep.Storer.NewEncodedObject(); false {
		//
	} else if tree_obj.SetType(plumbing.TreeObject); false {
		//
	} else if t, e := rep.Storer.SetEncodedObject(tree_obj); e != nil {
		err = e
	} else if commit_obj := rep.Storer.NewEncodedObject(); false {
		//
	} else if commit_obj.SetType(plumbing.CommitObject); false {
		//
	} else if w, e := commit_obj.Writer(); e != nil {
		err = e
	} else if timestring := fmt.Sprintf("%d %s", time.Now().Unix(), time.Now().Format("-0700")); false {
		//
	} else if _, e := fmt.Fprintf(w, "tree %s\n", t); e != nil {
		err = e
	} else if _, e := fmt.Fprintf(w, "author atompub-git-bot <atompub-git-bot@localhost> %s\n", timestring); e != nil {
		err = e
	} else if _, e := fmt.Fprintf(w, "committer atompub-git-bot <atompub-git-bot@localhost> %s\n\ninit commit\n", timestring); e != nil {
		err = e
	} else if e := w.Close(); e != nil {
		err = e
	} else {
		return rep.Storer.SetEncodedObject(commit_obj)
	}
	return
}

// store objects in the repository storer and populate hashmap
func (s *BillyStorer) storeObjs() error {
	for _, dir := range []string{"entry", "source"} {
		if entries, e := s.fsys.ReadDir(dir); e != nil {
			return e
		} else {
			for _, entry := range entries {
				fpath := path.Join(dir, entry.Name())
				if f, e := s.fsys.Open(fpath); e != nil {
					return e
				} else if obj := s.rep.Storer.NewEncodedObject(); false {
					//
				} else if obj.SetType(plumbing.BlobObject); false {
					//
				} else if w, e := obj.Writer(); e != nil {
					return e
				} else if _, e := io.Copy(w, f); e != nil {
					return e
				} else if e := w.Close(); e != nil {
					return e
				} else if e := f.Close(); e != nil {
					return e
				} else if e := s.fsys.Remove(fpath); e != nil {
					return e
				} else if h, e := s.rep.Storer.SetEncodedObject(obj); e != nil {
					return e
				} else {
					s.hashmap[dir][entry.Name()] = h
				}
			}
		}
	}
	return nil
}

// returns the hash of the next tree
func (s *BillyStorer) nextTree() (h plumbing.Hash, err error) {
	if h_entry, e := treeHelper(s.rep.Storer, s.hashmap["entry"]); e != nil {
		err = e
	} else if h_source, e := treeHelper(s.rep.Storer, s.hashmap["source"]); e != nil {
		err = e
	} else if h_root, e := func() (h plumbing.Hash, err error) {
		obj := s.rep.Storer.NewEncodedObject()
		obj.SetType(plumbing.TreeObject)
		w, e := obj.Writer()
		if e != nil {
			err = e
			return
		}
		names := [2][]byte{[]byte("entry"), []byte("source")}
		for k, v := range [2]plumbing.Hash{h_entry, h_source} {
			if _, e := w.Write([]byte(filemode.Dir.String())); e != nil {
				err = e
			} else if _, e := w.Write([]byte{' '}); e != nil {
				err = e
			} else if _, e := w.Write(names[k]); e != nil {
				err = e
			} else if _, e := w.Write([]byte{0x00}); e != nil {
				err = e
			} else if _, e := w.Write(v[:]); e != nil {
				err = e
			}
		}
		if err != nil {
			//
		} else if k, e := s.rep.Storer.SetEncodedObject(obj); e != nil {
			err = e
		} else {
			h = k
		}
		return
	}(); e != nil {
		err = e
	} else {
		h = h_root
	}
	return
}

func treeHelper(storer storer.Storer, hashmapchild map[string]plumbing.Hash) (h plumbing.Hash, err error) {
	obj := storer.NewEncodedObject()
	obj.SetType(plumbing.TreeObject)
	w, e := obj.Writer()
	if e != nil {
		err = e
		return
	}
	for k, v := range hashmapchild {
		if _, e := w.Write([]byte(filemode.Regular.String())); e != nil {
			err = e
		} else if _, e := w.Write([]byte{' '}); e != nil {
			err = e
		} else if _, e := w.Write([]byte(k)); e != nil {
			err = e
		} else if _, e := w.Write([]byte{0x00}); e != nil {
			err = e
		} else if _, e := w.Write(v[:]); e != nil {
			err = e
		}
	}
	if err != nil {
		//
	} else if k, e := storer.SetEncodedObject(obj); e != nil {
		err = e
	} else {
		h = k
	}
	return
}

func (s *BillyStorer) nextCommit(tree_hash plumbing.Hash, message string) (err error) {
	obj := s.rep.Storer.NewEncodedObject()
	obj.SetType(plumbing.CommitObject)

	if ref, e := s.rep.Reference(plumbing.Master, true); e != nil {
		err = e
	} else if parent_commit_obj, e := s.rep.CommitObject(ref.Hash()); e != nil {
		err = e
	} else if w, e := obj.Writer(); e != nil {
		err = e
	} else if timestring := fmt.Sprintf("%d %s", time.Now().Unix(), time.Now().Format("-0700")); false {
		//
	} else if _, e := fmt.Fprintf(w, "tree %s\n", tree_hash.String()); e != nil {
		err = e
	} else if _, e := fmt.Fprintf(w, "parent %s\n", ref.Hash().String()); e != nil {
		err = e
	} else if _, e := fmt.Fprintf(w, "author %s %s\n", parent_commit_obj.Author.String(), timestring); e != nil {
		err = e
	} else if _, e := fmt.Fprintf(w, "committer %s %s\n\n%s\n", parent_commit_obj.Committer.String(), timestring, message); e != nil {
		err = e
	} else if e := w.Close(); e != nil {
		err = e
	} else if h, e := s.rep.Storer.SetEncodedObject(obj); e != nil {
		err = e
	} else if ref := plumbing.NewHashReference(plumbing.Master, h); false {
		//
	} else if e := s.rep.Storer.SetReference(ref); e != nil {
		err = e
	}
	return

}
