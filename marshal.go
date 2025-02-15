package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"mime"
	"net/url"
	"time"
)

const atom_xmlns = `http://www.w3.org/2005/Atom`
const app_xmlns = `http://www.w3.org/2007/app`

func (f *Feed) MarshalTo(bw *bufio.Writer) (err error) {
	if e := f.Validate(); e != nil {
		return e
	}
	parent := xml.Name{Space: atom_xmlns, Local: "feed"}
	defer bw.WriteString(`</feed>`)
	if _, err = fmt.Fprintf(bw, "<feed xmlns=\"%s\">", atom_xmlns); err != nil {
	} else if err = f.Id.MarshalTo(bw); err != nil {
	} else if err = f.Title.MarshalTo(bw, parent); err != nil {
	} else if err = f.Subtitle.MarshalTo(bw, parent); err != nil {
	} else if err = f.Updated.MarshalTo(bw); err != nil {
	}

	if err != nil {
		return
	}

	for _, a := range f.Authors {
		if e := a.MarshalTo(bw); e != nil {
			return e
		}
	}

	for _, c := range f.Contributors {
		if e := c.MarshalTo(bw); e != nil {
			return e
		}
	}

	// categories
	for _, c := range f.Categories {
		if e := c.MarshalTo(bw, parent); e != nil {
			return e
		}
	}

	for _, l := range f.Links {
		if e := l.MarshalTo(bw); e != nil {
			return e
		}
	}

	if err = f.Icon.MarshalTo(bw); err != nil {
		//
	} else if err = f.Logo.MarshalTo(bw); err != nil {
		//
	} else if err = f.Rights.MarshalTo(bw, parent); err != nil {
		//
	} else if err = f.Generator.MarshalTo(bw); err != nil {
		//
	} else if err = f.Collection.MarshalTo(bw, parent); err != nil {
		//
	}
	if err != nil {
		return
	}

	// entries
	for _, t := range f.Entries {
		if e := t.MarshalTo(bw, f.Id); e != nil {
			return e
		}
	}

	return
}

func (s *Source) MarshalTo(bw *bufio.Writer, feed_id *URI, t *Entry) (err error) {
	if s == nil || feed_id.Consumes(s.Id) {
		return
	}
	parent := xml.Name{Space: atom_xmlns, Local: "source"}

	if t == nil {
		if _, err = fmt.Fprintf(bw, "<feed xmlns=\"%s\">", atom_xmlns); err != nil {
			return
		}
		defer bw.WriteString(`</feed>`)
	} else {
		if _, err = fmt.Fprintf(bw, "<source>"); err != nil {
			return
		}
		defer bw.WriteString(`</source>`)
	}

	if err = s.Id.MarshalTo(bw); err != nil {
		return
	}

	if err != nil {
		return
	}

	// marshalling the source on its own
	if t == nil {
		if err = s.Title.MarshalTo(bw, parent); err != nil {
			//
		} else if err = s.Subtitle.MarshalTo(bw, parent); err != nil {
			//
		} else if err = s.Updated.MarshalTo(bw); err != nil {
			//
		}
		if err != nil {
			return
		}

		for _, a := range s.Authors {
			if e := a.MarshalTo(bw); e != nil {
				return e
			}

		}

		for _, c := range s.Contributors {
			if e := c.MarshalTo(bw); e != nil {
				return e
			}
		}

		for _, l := range s.Links {
			if e := l.MarshalTo(bw); e != nil {
				return e
			}
		}

		for _, c := range s.Categories {
			if e := c.MarshalTo(bw, parent); e != nil {
				return e
			}
		}

		if err = s.Icon.MarshalTo(bw); err != nil {
			//
		} else if err = s.Logo.MarshalTo(bw); err != nil {
			//
		} else if err = s.Rights.MarshalTo(bw, parent); err != nil {
			//
		}
	} else {
		// marshalling the source in an event
		if t.Authors == nil {
			for _, a := range s.Authors {
				if e := a.MarshalTo(bw); e != nil {
					return e
				}
			}
		}
		if t.Rights == nil {
			if err = s.Rights.MarshalTo(bw, parent); err != nil {
				//
			}
		}
	}
	return
}

func (t *Entry) MarshalTo(bw *bufio.Writer, feed_id *URI) (err error) {
	if t == nil {
		return
	}

	var header string
	var footer string
	if feed_id != nil {
		header = "entry"
		footer = "entry"
	} else if _, e := t.Validate(feed_id); e != nil {
		return e
	} else {
		header = fmt.Sprintf("entry xmlns=\"%s\"", atom_xmlns)
		footer = "entry"
	}

	sub_parent := xml.Name{Space: atom_xmlns, Local: "entry"}
	defer fmt.Fprintf(bw, "</%s>", footer)
	if _, err = fmt.Fprintf(bw, "<%s>", header); err != nil {
	} else if err = t.Title.MarshalTo(bw, sub_parent); err != nil {
	} else if err = t.Summary.MarshalTo(bw, sub_parent); err != nil {
	} else if err = t.Updated.MarshalTo(bw); err != nil {
	} else if err = t.Published.MarshalTo(bw); err != nil {
	} else if err = t.Edited.MarshalTo(bw); err != nil {
	} else if err = t.Id.MarshalTo(bw); err != nil {
	}
	if err != nil {
		return
	}

	for _, a := range t.Authors {
		if e := a.MarshalTo(bw); e != nil {
			return e
		}
	}

	for _, c := range t.Contributors {
		if e := c.MarshalTo(bw); e != nil {
			return e
		}
	}

	if e := t.Content.MarshalTo(bw); e != nil {
		return e
	}

	for _, l := range t.Links {
		if e := l.MarshalTo(bw); e != nil {
			return e
		}
	}

	for _, c := range t.Categories {
		if e := c.MarshalTo(bw, sub_parent); e != nil {
			return e
		}
	}

	if err = t.Rights.MarshalTo(bw, sub_parent); err != nil {
		//
	} else if err = t.Control.MarshalTo(bw); err != nil {
		//
	} else if err = t.Source.MarshalTo(bw, feed_id, t); err != nil {
		//
	}
	return
}

func (p *Person) MarshalTo(bw *bufio.Writer) (err error) {
	if p == nil {
		return
	}
	var tag string
	switch p.XMLName.Space {
	case atom_xmlns:
		tag = p.XMLName.Local
	default:
		return fmt.Errorf("unknown Person xmlns")
	}

	if p.Name == "" {
		if _, err = fmt.Fprintf(bw, "<%s><name/>", tag); err != nil {
			return
		}
	} else if _, err = fmt.Fprintf(bw, "<%s><name>%s</name>", tag, p.Name); err != nil {
		return
	}

	if err = p.URI.MarshalTo(bw); err != nil {
		return
	}

	_, err = fmt.Fprintf(bw, "</%s>", tag)

	return
}

func (t *Content) MarshalTo(bw *bufio.Writer) (err error) {
	if t == nil {
		// if no content, require link
		return
	}

	if t.Src != "" {
		// out of line content
		if _, err = fmt.Fprintf(bw, "<content src=\"%s\"", t.Src); err != nil {
			//
		} else if t.Type == "" {
			_, err = bw.WriteString("/>")
		} else if _, err = fmt.Fprintf(bw, " type=\"%s\"/>", t.Type); err != nil {
			//
		}
		return
	}
	// in line content

	if _, err = fmt.Fprintf(bw, "<content type=\"%s\">", t.Type); err != nil {
		return
	} else if _, err = bw.Write(t.Body); err != nil {
		return
	} else if _, err = bw.WriteString("</content>"); err != nil {
		return
	}
	return
}

func (t *TextConstruct) MarshalTo(bw *bufio.Writer, parent xml.Name) (err error) {
	if t == nil {
		return
	}
	var header string
	var footer string
	switch t.XMLName.Space {
	case atom_xmlns:
		if parent.Space != atom_xmlns {
			header = "atom:" + t.XMLName.Local
			footer = "atom:" + t.XMLName.Local
		} else {
			header = t.XMLName.Local
			footer = t.XMLName.Local
		}
	default:
		return fmt.Errorf("unknown text construct xmlns")
	}
	if t.Type == "" {
		t.Type = "text"
	}

	if _, err = fmt.Fprintf(bw, "<%s", header); err != nil {
		//
	} else if t.Text != "" {
		_, err = fmt.Fprintf(bw, " type=\"%s\">%s</%s>", t.Type, t.Text, footer)
	} else {
		_, err = fmt.Fprintf(bw, "/>")
	}

	return
}

func (t *DateConstruct) MarshalTo(bw *bufio.Writer) (err error) {
	if t == nil {
		return
	}
	var header, footer string
	switch t.XMLName.Space {
	case atom_xmlns: // updated, published
		header = t.XMLName.Local
		footer = t.XMLName.Local
	case app_xmlns: // edited
		header = fmt.Sprintf(`%s xmlns="%s"`, t.XMLName.Local, app_xmlns)
		footer = t.XMLName.Local
	default:
		return fmt.Errorf("unknown date construct xmlns")
	}
	_, err = fmt.Fprintf(bw, "<%s>%s</%s>", header, time.Time(t.T).Format(time.RFC3339Nano), footer)
	return
}

func (u *URI) MarshalTo(bw *bufio.Writer) (err error) {
	if u == nil {
		return
	}
	var tag string
	switch u.XMLName.Space {
	case atom_xmlns:
		tag = u.XMLName.Local
	default:
		return fmt.Errorf("unknown URI construct xmlns")
	}
	if _, err = url.Parse(u.Target); err != nil {
		//
	} else if _, err = fmt.Fprintf(bw, "<%s>%s</%s>", tag, u.Target, tag); err != nil {
		//
	}
	return
}

func (l *Link) MarshalTo(bw *bufio.Writer) (err error) {
	if l == nil {
		return
	}
	if _, err = bw.WriteString("<link"); err != nil {
		return
	}

	defer bw.WriteString("/>")

	if _, err = url.Parse(l.Href); err != nil {
		return
	} else if _, err = fmt.Fprintf(bw, " href=\"%s\"", l.Href); err != nil {
		return
	}

	if l.Relation == "" {
		//
	} else if _, err = fmt.Fprintf(bw, " rel=\"%s\"", l.Relation); err != nil {
		return
	}

	if l.Type == "" {
		//
	} else if _, _, err = mime.ParseMediaType(l.Type); err != nil {
		return
	} else if _, err = fmt.Fprintf(bw, " type=\"%s\"", l.Type); err != nil {
		return
	}

	if l.HrefLang == "" {
		//
	} else if _, err = fmt.Fprintf(bw, " hreflang=\"%s\"", l.HrefLang); err != nil {
		return
	}

	if l.Title == "" {
		//
	} else if _, err = fmt.Fprintf(bw, " title=\"%s\"", l.Title); err != nil {
		return
	}

	if l.Length == 0 {
		//
	} else if _, err = fmt.Fprintf(bw, " length=\"%d\"", l.Length); err != nil {
		return
	}

	return
}

func (c *Category) MarshalTo(bw *bufio.Writer, parent xml.Name) (err error) {
	if c == nil {
		return
	}

	var header string
	switch parent.Space {
	case atom_xmlns:
		header = "category"
	case app_xmlns:
		header = "atom:category"
	default:
		err = fmt.Errorf("unknown parent xmlns for category")
		return
	}

	if _, err = fmt.Fprintf(bw, "<%s", header); err != nil {
		return
	}

	defer bw.WriteString("/>")

	if c.Term == "" {
		return fmt.Errorf("Category is missing term")
	} else if _, err = fmt.Fprintf(bw, " term=\"%s\"", c.Term); err != nil {
		return
	}

	if scheme := c.Scheme; scheme == "" {
		// todo: check is: alternate, related, self, enclosure, via
	} else if _, err = url.Parse(scheme); err != nil {
		return
	} else if _, err = fmt.Fprintf(bw, " scheme=\"%s\"", scheme); err != nil {
		return
	}

	if c.Label == "" {
		//
	} else if _, err = fmt.Fprintf(bw, " label=\"%s\"", c.Label); err != nil {
		return
	}

	return
}

func (g *Generator) MarshalTo(bw *bufio.Writer) (err error) {
	if g == nil {
		return
	}
	if _, err = bw.WriteString("<generator"); err != nil {
		return
	}

	if g.URI == "" {
		//
	} else if _, err = url.Parse(g.URI); err != nil {
		return
	} else if _, err = fmt.Fprintf(bw, " uri=\"%s\"", g.URI); err != nil {
		return
	}

	if g.Version == "" {
		//
	} else if _, err = fmt.Fprintf(bw, " version=\"%s\"", g.Version); err != nil {
		return
	}

	if text := g.Text; text != "" {
		_, err = fmt.Fprintf(bw, ">%s</generator>", text)
	} else {
		fmt.Fprintf(bw, "/>")
	}

	return
}

// app
func (s *Service) MarshalTo(bw *bufio.Writer) (err error) {
	header := `<service xmlns:atom="http://www.w3.org/2005/Atom" xmlns="http://www.w3.org/2007/app">`
	footer := "</service>"
	if len(s.Workspaces) == 0 {
		err = fmt.Errorf("service document requires at least one workspace")
		return
	}
	if _, err = bw.WriteString(header); err != nil {
		return
	}
	for _, w := range s.Workspaces {
		if err = w.MarshalTo(bw); err != nil {
			return
		}
	}
	_, err = bw.WriteString(footer)
	return
}

func (w *Workspace) MarshalTo(bw *bufio.Writer) (err error) {
	if w == nil {
		return
	}
	parent := xml.Name{Space: app_xmlns, Local: "workspace"}
	if _, err = bw.WriteString("<workspace>"); err != nil {
		//
	} else if err = w.Title.MarshalTo(bw, parent); err != nil {
		//
	} else {
		for _, c := range w.Collections {
			if err = c.MarshalTo(bw, parent); err != nil {
				return
			}
		}
	}
	if err != nil {
		return
	}
	_, err = bw.WriteString("</workspace>")
	return
}

func (c *Collection) MarshalTo(bw *bufio.Writer, parent xml.Name) (err error) {
	if c == nil {
		return
	}
	var header, footer string
	switch parent.Space {
	case app_xmlns:
		header = "collection"
		footer = "collection"
	case atom_xmlns:
		header = `collection xmlns:atom="http://www.w3.org/2005/Atom" xmlns="http://www.w3.org/2007/app"`
		footer = "collection"
	default:
		err = fmt.Errorf("unknown parent xmlns for app:collection")
	}

	sub_parent := xml.Name{Space: app_xmlns, Local: "collection"}
	if c.Href == "" {
		err = fmt.Errorf("empty collection href")
	} else if _, err = url.Parse(c.Href); err != nil {
		//
	} else if _, err = fmt.Fprintf(bw, "<%s href=\"%s\">", header, c.Href); err != nil {
		//
	} else if err = c.Title.MarshalTo(bw, sub_parent); err != nil {
		//
	}
	if err != nil {
		return
	}
	for _, cat := range c.Categories {
		if err = cat.MarshalTo(bw); err != nil {
			return
		}
	}
	for _, a := range c.Accepts {
		if err = a.MarshalTo(bw); err != nil {
			return
		}
	}
	_, err = fmt.Fprintf(bw, "</%s>", footer)
	return
}

func (c *Categories) MarshalTo(bw *bufio.Writer) (err error) {
	if c == nil {
		return
	}
	if c.Href != "" {
		// out of line
		if _, e := url.Parse(c.Href); e != nil {
			err = e
			return
		}
		_, err = fmt.Fprintf(bw, "<categories href=\"%s\"/>", c.Href)
		return
	}
	// in line
	switch c.Fixed {
	case "no", "":
		if _, err = bw.WriteString("<categories"); err != nil {
			return
		}
	case "yes":
		if _, err = bw.WriteString("<categories fixed=\"yes\""); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unknown fixed attribute")
		return
	}
	if c.Scheme != "" {
		if _, e := url.Parse(c.Scheme); e != nil {
			err = e
			return
		}
		if _, err = fmt.Fprintf(bw, " scheme=\"%s\">", c.Scheme); err != nil {
			return
		}
	} else {
		if _, err = bw.WriteString(">"); err != nil {
			return
		}
	}
	parent := xml.Name{Space: app_xmlns, Local: "categories"}
	for _, cat := range c.Categories {
		if err = cat.MarshalTo(bw, parent); err != nil {
			return
		}
	}
	_, err = bw.WriteString("</categories>")
	return
}

func (a *Accept) MarshalTo(bw *bufio.Writer) (err error) {
	if a == nil {
		return
	}
	if a.Text == "" {
		_, err = fmt.Fprintf(bw, "<accept/>")
		return
	}
	_, err = fmt.Fprintf(bw, "<accept>%s</accept>", a.Text)
	return
}

func (c *PublishingControl) MarshalTo(bw *bufio.Writer) (err error) {
	if c == nil || c.Draft != "yes" {
		return
	}

	_, err = fmt.Fprintf(bw, `<control xmlns="%s"><draft>yes</draft></control>`, app_xmlns)
	return
}
