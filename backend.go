package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Backend struct {
	serviceDocument *Service
	sourcemap       map[string]*Source
	entrymap        map[string]*Entry
	storer          Storer
}

func NewBackend(storer Storer) *Backend {
	b := &Backend{
		serviceDocument: &Service{
			Workspaces: []Workspace{
				{
					Title: TextConstruct{
						XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
						Text:    "default workspace",
					},
					Collections: make([]*Collection, 0, 32),
				},
			},
		},
		entrymap:  make(map[string]*Entry),
		sourcemap: make(map[string]*Source),
		storer:    storer,
	}
	if e := storer.Populate(b.entrymap, b.sourcemap); e != nil {
		panic(e)
	}
	// now build the serviceDocument
	for k, v := range b.sourcemap {
		if v.Title == nil || v.Title.XMLName.Space != atom_xmlns || v.Title.XMLName.Local != "title" {
			panic("invalid source title")
		}
		collection := &Collection{
			Href:  k,
			Title: v.Title,
			Accepts: []Accept{{
				Text: "text/plain",
			}},
			Categories: []Categories{{
				Categories: v.Categories,
			}},
		}
		v.Collection = collection
		b.serviceDocument.Workspaces[0].Collections = append(b.serviceDocument.Workspaces[0].Collections, collection)
	}

	return b
}

func (b *Backend) GetRoot(r *http.Request) (sd *Service, err *HTTPError) {
	return b.serviceDocument, nil
}

func (b *Backend) PostToRoot(r *http.Request, new_feed *Feed) (feed *Feed, feed_URL string, err *HTTPError) {
	feed = new_feed
	if new_feed.Title == nil || new_feed.Authors == nil {
		err = &HTTPError{code: http.StatusBadRequest, message: "need to set <title> and <author> for new feeds"}
		return
	}

	var uuid_string string

	if new_feed.Id == nil {
		uuid_string = uuid.NewString()
	} else if u, e := uuid.Parse(new_feed.Id.Target); e != nil {
		uuid_string = uuid.NewString()
	} else {
		uuid_string = u.String()
	}

	feed_URL = "/feed/" + uuid_string
	if _, ok := b.sourcemap[feed_URL]; ok {
		err = &HTTPError{code: http.StatusConflict, message: "feed with given URI already exists"}
		return
	}

	new_feed.Id = &URI{
		XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
		Target:  "urn:uuid:" + uuid_string,
	}

	new_feed.Updated = &DateConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
		T:       time.Now().Round(time.Microsecond),
	}

	new_feed.Collection = &Collection{
		Href:  feed_URL,
		Title: new_feed.Title,
		Accepts: []Accept{
			{
				Text: "text/plain",
			},
		},
		Categories: []Categories{{
			Categories: new_feed.Categories,
		}},
	}

	if e := new_feed.Validate(); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: e.Error()}
		return
	}

	source := &Source{
		Id:           new_feed.Id,
		Updated:      new_feed.Updated,
		Authors:      new_feed.Authors,
		Title:        new_feed.Title,
		Links:        new_feed.Links,
		Categories:   new_feed.Categories,
		Icon:         new_feed.Icon,
		Logo:         new_feed.Logo,
		Subtitle:     new_feed.Subtitle,
		Rights:       new_feed.Rights,
		Contributors: new_feed.Contributors,
		Collection:   new_feed.Collection,
	}

	b.sourcemap[feed_URL] = source

	if e := b.storer.AddSource(source); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		return
	} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		return
	}

	for k, v := range b.serviceDocument.Workspaces[0].Collections {
		if v == nil {
			b.serviceDocument.Workspaces[0].Collections[k] = new_feed.Collection
			return
		}
	}

	b.serviceDocument.Workspaces[0].Collections = append(b.serviceDocument.Workspaces[0].Collections, new_feed.Collection)
	return
}

func (b *Backend) PutFeed(r *http.Request, new_feed *Feed) (err *HTTPError) {
	if v, ok := b.sourcemap[r.URL.Path]; !ok {
		err = &HTTPError{code: http.StatusNotFound}
	} else if !v.Id.Consumes(new_feed.Id) {
		return &HTTPError{code: http.StatusBadRequest, message: "cannot change the URI of the feed"}
	} else {
		v.Collection.Title = new_feed.Title
		v.Collection.Categories = []Categories{{
			Categories: new_feed.Categories,
		}}
		v.Authors = new_feed.Authors
		v.Contributors = new_feed.Contributors
		v.Title = new_feed.Title
		v.Subtitle = new_feed.Subtitle
		v.Icon = new_feed.Icon
		v.Logo = new_feed.Logo
		v.Links = new_feed.Links
		v.Rights = new_feed.Rights
		v.Categories = new_feed.Categories

		//
		v.Updated.Set(time.Now().Round(time.Microsecond))

		if e := b.storer.AddSource(v); e != nil {
			err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
			err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		}
	}
	return
}

func (b *Backend) DeleteFeed(r *http.Request) (err *HTTPError) {
	if source, ok := b.sourcemap[r.URL.Path]; !ok {
		err = &HTTPError{code: http.StatusNotFound}
	} else if source == nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
	} else if e := b.storer.DeleteSource(source); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
	} else {
		delete(b.sourcemap, r.URL.Path)
		for k, v := range b.entrymap {
			if v == nil || v.Source == nil || v.Source.Id == nil {
				err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
				return
			} else if !source.Id.Consumes(v.Source.Id) {
				// continue
			} else if e := b.storer.DeleteEntry(v); e != nil {
				err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
			} else {
				delete(b.entrymap, k)
			}
		}
		for k, v := range b.serviceDocument.Workspaces[0].Collections {
			if v == nil {
				continue
			} else if u, e := url.Parse(v.Href); e != nil {
				err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
				return
			} else if u.Path == r.URL.Path {
				b.serviceDocument.Workspaces[0].Collections[k] = nil
			}
		}

	}
	if err != nil {
		return
	} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	}
	return
}

func (b *Backend) GetFeed(r *http.Request) (feed *Feed, err *HTTPError) {
	source, ok := b.sourcemap[r.URL.Path]
	if !ok {
		err = &HTTPError{code: http.StatusNotFound}
		return
	} else if source == nil || source.Id == nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
		return
	}

	// collect the entries
	entry_ptrs := make([]*Entry, 0, 128)
	for _, v := range b.entrymap {
		if v == nil || v.Source == nil {
			err = &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
			return
		} else if !source.Id.Consumes(v.Source.Id) {
			// continue
		} else {
			entry_ptrs = append(entry_ptrs, v)
		}
	}

	// sort the entries
	sort.Slice(entry_ptrs, func(i int, j int) bool {
		// do not check error; the nil pointers will be sorted last
		ti, _ := entry_ptrs[i].UpdatedTime()
		tj, _ := entry_ptrs[j].UpdatedTime()
		return ti.After(tj)
	})

	return &Feed{
		Id:         source.Id,
		Authors:    source.Authors,
		Updated:    source.Updated,
		Rights:     source.Rights,
		Links:      source.Links,
		Title:      source.Title,
		Subtitle:   source.Subtitle,
		Icon:       source.Icon,
		Logo:       source.Logo,
		Categories: source.Categories,
		Entries:    entry_ptrs,
	}, nil
}

func (b *Backend) PostToFeed(r *http.Request) (entry *Entry, entry_URL string, err *HTTPError) {
	var source *Source
	if sd := b.serviceDocument; false {
		//
	} else {
		for _, w := range sd.Workspaces {
			for _, c := range w.Collections {
				if c == nil {
					continue
				}
				if u, e := url.Parse(c.Href); e != nil {
					panic(e)
				} else if u.Path == r.URL.Path {
					// found the correct collection
					source = b.sourcemap[r.URL.Path]
					for _, a := range c.Accepts {
						if strings.EqualFold(r.Header.Get("Content-Type"), a.Text) {
							goto jump
						}
					}
					err = &HTTPError{code: http.StatusUnsupportedMediaType}
					return
				}
			}
		}
		err = &HTTPError{code: http.StatusNotFound}
		return
	}

jump:
	var uuid_string string
	var cats []string
	switch r.Header.Get("Content-Type") {
	case "text/plain":
		entry, cats, uuid_string, err = plainTextPost(r.Body, r.Header.Get("Slug"))
	default:
		err = &HTTPError{code: http.StatusInternalServerError, message: "backend only accepts text/plain"}
	}

	entry.Source = source
	entry.Categories = make([]Category, 0, len(cats))
	for _, term := range cats {
		for _, cat := range source.Categories {
			if cat.Term == term {
				entry.Categories = append(entry.Categories, cat)
			}
		}
	}

	if err != nil {
		return
	} else if _, e := entry.Validate(nil); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: e.Error()}
		return
	} else if e := source.Updated.Set(time.Now().Round(time.Microsecond)); e != nil {
		panic(e)
	} else if e := b.storer.AddEntry(entry); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		return
	} else if e := b.storer.AddSource(source); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		return
	} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
		err = &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		return
	}

	entry_relative := "/entry/" + uuid_string
	b.entrymap[entry_relative] = entry
	entry_URL = entry_relative

	return
}

var protocol_list = []string{"cat", "http", "https", "mailto", "tel", "sms"}

func preparePlainText(input io.Reader) (output []byte, categories []string, err error) {
	buf := bytes.NewBufferString(`<div xmlns="http://www.w3.org/1999/xhtml" style="white-space: pre-line;"><p>`)
	urls := make([]string, 0, 8)
	categories = make([]string, 0, 8)
	input_bytes, e := io.ReadAll(input)
	if e != nil {
		err = e
		return
	}
	for {
		before, after, found := bytes.Cut(input_bytes, []byte{':'})
		if !found {
			if e := xml.EscapeText(buf, before); e != nil {
				err = e
				return
			} else {
				break
			}
		} else if i := bytes.IndexAny(after, "\n "); i != -1 {
			input_bytes = after[i:]
			after = after[:i]
		} else {
			input_bytes = nil
		}
		for k, v := range protocol_list {
			if bytes.HasSuffix(before, []byte(v)) {
				raw_url := fmt.Sprintf("%s:%s", v, after)
				if _, e := url.Parse(raw_url); e == nil {
					before = before[:len(before)-len(v)]
					if k == 0 {
						// category!
						categories = append(categories, strings.TrimPrefix(raw_url, "cat:"))
						// trim the next char since we are not inserting anything to replace cat:XXX
						if input_bytes != nil {
							input_bytes = input_bytes[1:]
						}
						goto jump
					}

					// often want to put a . after the url reference
					if len(input_bytes) > 1 && input_bytes[0] == ' ' && input_bytes[1] == '.' {
						input_bytes = input_bytes[1:]
					}

					for l, u := range urls {
						if u == raw_url {
							before = append(before, []byte(fmt.Sprintf("[%d]", l+1))...)
							goto jump
						}
					}
					urls = append(urls, raw_url)
					before = append(before, []byte(fmt.Sprintf("[%d]", len(urls)))...)
					goto jump
				}
			}
		}
		before = append(before, ':')
		before = append(before, after...)
	jump:
		if e := xml.EscapeText(buf, before); e != nil {
			err = e
			return
		}
		if input_bytes == nil {
			break
		}
	}
	if _, e := buf.WriteString("</p>"); e != nil {
		err = e
		return
	}
	if len(urls) != 0 {
		tmp := bytes.NewBuffer(nil)
		if _, e := fmt.Fprintf(buf, `<div style="word-break:break-all;">`); e != nil {
			err = e
			return
		}
		for k, u := range urls {
			if e := xml.EscapeText(tmp, []byte(u)); e != nil {
				err = e
				return
			} else if _, e := fmt.Fprintf(buf, `<a href="%s">[%d]&#xA0;%s</a>`, tmp.String(), k+1, tmp.String()); e != nil {
				err = e
				return
			}
			tmp.Reset()
		}
		if _, e := fmt.Fprintf(buf, `</div>`); e != nil {
			err = e
			return
		}
	}

	_, err = buf.WriteString("</div>")
	output = buf.Bytes()
	return
}

func plainTextPost(body io.Reader, slug string) (entry *Entry, cats []string, uuid_string string, err *HTTPError) {
	entry = &Entry{}

	if output, categories, e := preparePlainText(body); e != nil {
		err = &HTTPError{code: http.StatusBadRequest, message: e.Error()}
		return
	} else {
		cats = categories
		entry.Content.Type = "xhtml"
		entry.Content.Body = output
	}

	if slug == "" {
		slug = "Untitled"
	}

	entry.Title = TextConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "title"},
		Text:    slug,
	}

	uuid_string = uuid.NewString()

	entry.Updated = DateConstruct{
		XMLName: xml.Name{Space: atom_xmlns, Local: "updated"},
		T:       time.Now().Round(time.Second),
	}

	entry.Id = URI{
		XMLName: xml.Name{Space: atom_xmlns, Local: "id"},
		Target:  "urn:uuid:" + uuid_string,
	}

	return
}

func (b *Backend) GetEntry(r *http.Request) (entry *Entry, err *HTTPError) {
	if ent, ok := b.entrymap[r.URL.Path]; !ok {
		err = &HTTPError{code: http.StatusNotFound}
		return
	} else {
		return ent, nil
	}
}

func (b *Backend) DeleteEntry(r *http.Request) (err *HTTPError) {
	if entry, ok := b.entrymap[r.URL.Path]; !ok {
		return &HTTPError{code: http.StatusNotFound}
	} else if entry.Source == nil {
		return &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
	} else if e := entry.Source.Updated.Set(time.Now().Round(time.Microsecond)); e != nil {
		return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else if e := b.storer.DeleteEntry(entry); e != nil {
		return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
		return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else {
		delete(b.entrymap, r.URL.Path)
	}
	return
}

func (b *Backend) PutEntry(r *http.Request, new_entry *Entry) (err *HTTPError) {
	if entry, ok := b.entrymap[r.URL.Path]; !ok {
		return &HTTPError{code: http.StatusNotFound}
	} else if entry.Source == nil || entry.Source.Updated == nil {
		return &HTTPError{code: http.StatusInternalServerError, message: "nil pointer dereference"}
	} else if !entry.Id.Consumes(&new_entry.Id) {
		return &HTTPError{code: http.StatusBadRequest, message: "cannot change the URI of the entry"}
	} else if e := entry.Updated.Set(time.Now().Round(time.Second)); e != nil {
		return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else if e := entry.Source.Updated.Set(time.Now().Round(time.Microsecond)); e != nil {
		return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
	} else {
		// note there was already a check in the handler by calling Validate()

		if entry.Title.Type != "text" && entry.Title.Type != "" {
			return &HTTPError{code: http.StatusBadRequest, message: "title type must be \"text\""}
		} else {
			entry.Title = new_entry.Title
		}

		if entry.Summary != nil && entry.Summary.Type != "text" && entry.Summary.Type != "" {
			return &HTTPError{code: http.StatusBadRequest, message: "summary type must be \"text\""}
		} else {
			entry.Summary = new_entry.Summary
		}

		if entry.Rights != nil && entry.Rights.Type != "text" && entry.Rights.Type != "" {
			return &HTTPError{code: http.StatusBadRequest, message: "rights type must be \"text\""}
		} else {
			entry.Rights = new_entry.Rights
		}

		entry.Authors = new_entry.Authors
		entry.Contributors = new_entry.Contributors
		entry.Control = new_entry.Control
		entry.Links = new_entry.Links
		entry.Categories = new_entry.Categories

		if new_entry.Content.Type != "xhtml" {
			return &HTTPError{code: http.StatusBadRequest, message: "cannot change content type from xhtml"}
		} else if new_entry.Content.Src != "" {
			return &HTTPError{code: http.StatusBadRequest, message: "content must be inline"}
		} else if new_entry.Content.Body, e = preparePutContent(new_entry.Content.Body); e != nil {
			return &HTTPError{code: http.StatusBadRequest, message: e.Error()}
		}

		entry.Content = new_entry.Content
		if e := b.storer.AddEntry(entry); e != nil {
			return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		} else if e := b.storer.AddSource(entry.Source); e != nil {
			return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		} else if e := b.storer.Commit(fmt.Sprintf("%s %s", r.Method, r.URL.Path)); e != nil {
			return &HTTPError{code: http.StatusInternalServerError, message: e.Error()}
		}
	}

	return
}

func preparePutContent(innerxml []byte) (prepared []byte, err error) {
	// remove newlines
	innerxml = bytes.Map(func(r rune) rune {
		if r == '\n' {
			return -1
		} else {
			return r
		}
	}, innerxml)
	dec := xml.NewDecoder(bytes.NewReader(innerxml))
	if t, e := dec.Token(); e != nil {
		err = e
	} else if tok, ok := t.(xml.StartElement); !ok {
		err = fmt.Errorf("content does not start with xhtml div")
	} else if tok.Name.Space != "http://www.w3.org/1999/xhtml" || tok.Name.Local != "div" {
		err = fmt.Errorf("content does not start with xhtml div")
	}
	if err != nil {
		return
	}
	for {
		if t, e := dec.Token(); errors.Is(e, io.EOF) {
			break
		} else if tok, ok := t.(xml.StartElement); !ok {
			continue
		} else {
			switch tok.Name.Local {
			case "p", "a":
				// allowed
			default:
				err = fmt.Errorf("tag %s not allowed", tok.Name.Local)
				return
			}
		}
	}
	prepared = innerxml
	return
}

func (b *Backend) GetMedia(r *http.Request) (media []byte, mediatype string, err *HTTPError) {
	err = &HTTPError{code: http.StatusNotImplemented}
	return
}
