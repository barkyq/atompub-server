package main

import (
	"fmt"
	"mime"
	"net/url"
	"strings"
)

func (f *Feed) Validate() error {
	if f == nil {
		return nil
	}
	if f.Title == nil || f.Title.Text == "" {
		return fmt.Errorf("error: empty title")
	} else if f.Updated == nil || f.Updated.XMLName.Local != "updated" || f.Updated.XMLName.Space != atom_xmlns {
		return fmt.Errorf("error: bad updated element")
	} else if f.Id == nil || f.Id.XMLName.Local != "id" || f.Id.XMLName.Space != atom_xmlns {
		return fmt.Errorf("error: bad id element")
	} else if f.Icon != nil && (f.Icon.XMLName.Local != "icon" || f.Icon.XMLName.Space != atom_xmlns) {
		return fmt.Errorf("error: bad icon element")
	} else if f.Logo != nil && (f.Logo.XMLName.Local != "logo" || f.Logo.XMLName.Space != atom_xmlns) {
		return fmt.Errorf("error: bad logo element")
	} else if f.Rights != nil && (f.Rights.XMLName.Local != "rights" || f.Rights.XMLName.Space != atom_xmlns) {
		return fmt.Errorf("error: bad rights element")
	} else if f.Subtitle != nil && (f.Subtitle.XMLName.Local != "subtitle" || f.Subtitle.XMLName.Space != atom_xmlns) {
		return fmt.Errorf("error: bad subtitle element")
	} else if f.Title.XMLName.Local != "title" || f.Title.XMLName.Space != atom_xmlns {
		return fmt.Errorf("error: bad title element")
	}

	var needs_author bool
	for _, t := range f.Entries {
		if ha, e := t.Validate(f.Id); e != nil {
			return e
		} else if !ha {
			needs_author = true
		}
	}
	if needs_author && len(f.Authors) == 0 {
		return fmt.Errorf("error: feed needs author")
	}
	for _, a := range f.Authors {
		if a.XMLName.Local != "author" || a.XMLName.Space != atom_xmlns {
			return fmt.Errorf("error: bad author xml name")
		}
	}
	for _, a := range f.Contributors {
		if a.XMLName.Local != "contributor" || a.XMLName.Space != atom_xmlns {
			return fmt.Errorf("error: bad contributor xml name")
		}
	}

	var needs_self = true
	for _, l := range f.Links {
		if rel, e := l.Validate(); e != nil {
			return e
		} else if rel == "self" {
			needs_self = false
		}
	}
	if needs_self {
		return fmt.Errorf("needs link with self relation")
	}

	return nil
}

func (s *Source) Validate(feed_id *URI) (has_author bool, err error) {
	if s == nil || feed_id.Consumes(s.Id) {
		return
	}
	if n := s.Updated; n != nil && (n.XMLName.Local != "updated" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad updated xml element")
	} else if n := s.Id; n == nil || n.XMLName.Local != "id" || n.XMLName.Space != atom_xmlns {
		// opinionated: need an ID
		err = fmt.Errorf("error: bad id xml element")
	} else if n := s.Icon; n != nil && (n.XMLName.Local != "icon" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad icon xml name")
	} else if n := s.Logo; n != nil && (n.XMLName.Local != "logo" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad logo xml name")
	} else if n := s.Rights; n != nil && (n.XMLName.Local != "rights" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad rights xml name")
	} else if n := s.Subtitle; n != nil && (n.XMLName.Local != "subtitle" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad subtitle xml name")
	} else if n := s.Title; n != nil && (n.XMLName.Local != "title" || n.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad title xml name")
	}
	if err != nil {
		return
	}

	for _, a := range s.Authors {
		if a.XMLName.Local != "author" || a.XMLName.Space != atom_xmlns {
			err = fmt.Errorf("error: bad author xml name")
			return
		}
		has_author = true
	}
	for _, a := range s.Contributors {
		if a.XMLName.Local != "contributor" || a.XMLName.Space != atom_xmlns {
			err = fmt.Errorf("error: bad contributor xml name")
			return
		}
	}

	for _, l := range s.Links {
		if _, e := l.Validate(); e != nil {
			err = e
			return
		}
	}
	return
}

func (l *Link) Validate() (relation string, err error) {
	if l == nil {
		err = fmt.Errorf("nil link")
		return
	}
	if l.Href == "" {
		err = fmt.Errorf("empty href")
		return
	}
	if _, err = url.Parse(l.Href); err != nil {
		return
	}

	switch l.Relation {
	case "":
		relation = "alternate"
	case "self", "related", "alternate", "enclosure", "via", "edit", "edit-media":
		relation = l.Relation
	default:
		err = fmt.Errorf("unknown link relation")
	}
	return
}

func (t *Entry) Validate(feed_id *URI) (has_author bool, err error) {
	if t == nil {
		return
	}

	if t.Title.Text == "" {
		err = fmt.Errorf("error: empty title")
	} else if t.Updated.XMLName.Local != "updated" || t.Updated.XMLName.Space != atom_xmlns {
		err = fmt.Errorf("error: bad updated xml name")
	} else if t.Id.XMLName.Local != "id" || t.Id.XMLName.Space != atom_xmlns {
		err = fmt.Errorf("error: bad id xml name")
	} else if t.Published != nil && (t.Published.XMLName.Local != "published" || t.Published.XMLName.Space != atom_xmlns) {
		err = fmt.Errorf("error: bad published xml name")
	} else if t.Edited != nil && (t.Edited.XMLName.Local != "edited" || t.Edited.XMLName.Space != app_xmlns) {
		err = fmt.Errorf("error: bad edited xml name")
	} else if has_author, err = t.Source.Validate(feed_id); err != nil {
		// propagate error
	} else if nal, ns, e := t.Content.Validate(); e != nil {
		err = e
	} else {
		if t.Summary == nil && ns {
			err = fmt.Errorf("needs summary")
			return
		}

		for _, l := range t.Links {
			if rel, e := l.Validate(); e != nil {
				err = e
				return
			} else if rel == "alternate" {
				nal = false
			}
		}
		if nal {
			err = fmt.Errorf("needs link with rel=alternate")
			return
		}
		for _, a := range t.Authors {
			if a.XMLName.Local != "author" || a.XMLName.Space != atom_xmlns {
				err = fmt.Errorf("error: bad author xml name")
				return
			}
			has_author = true
		}
		for _, a := range t.Contributors {
			if a.XMLName.Local != "contributor" || a.XMLName.Space != atom_xmlns {
				err = fmt.Errorf("error: bad contributor xml name")
				return
			}
		}
	}
	if feed_id == nil && !has_author {
		err = fmt.Errorf("error: entry needs author")
	}
	return
}

func (c *Content) Validate() (needs_alternate_link bool, needs_summary bool, err error) {
	if c == nil {
		needs_alternate_link = true
		return
	}
	if c.Src != "" {
		// out of line content
		needs_summary = true
		if _, err = url.Parse(c.Src); err != nil {
			return
		} else if _, _, err = mime.ParseMediaType(c.Type); err != nil {
			return
		}
		if c.Body != nil {
			err = fmt.Errorf("body must be empty if src is set")
			return
		}
	} else {
		// in line content
		switch c.Type {
		case "", "text", "html", "xhtml":
			return
		default:
			if mt, _, e := mime.ParseMediaType(c.Type); e != nil {
				err = e
				return
			} else if strings.HasPrefix(mt, "text/") {
				return
			} else if strings.HasSuffix(mt, "/xml") || strings.HasSuffix(mt, "+xml") {
				return
			} else {
				// should be base64 encoded
				needs_summary = true
				return
			}
		}
	}
	return
}
