package docxlib

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// unpack receives a zip file (word documents are a zip with multiple xml inside)
// and parses the files that are relevant for us:
//
//  1. Document
//  2. Relationships
func unpack(zipReader *zip.Reader) (docx *Docx, err error) {
	docx = new(Docx)
	docx.mediaNameIdx = make(map[string]int, 64)
	for _, f := range zipReader.File {
		if f.Name == "word/_rels/document.xml.rels" {
			err = docx.parseDocRelation(f)
			if err != nil {
				return
			}
		}
		if f.Name == "word/document.xml" {
			err = docx.parseDocument(f)
			if err != nil {
				return
			}
		}
		err = docx.checkAndParseMedia(f)
		if err != nil {
			return
		}
	}
	docx.buf = bytes.NewBuffer(make([]byte, 0, 1024*1024*4))
	return
}

// parseDocument processes one of the relevant files, the one with the actual document
func (f *Docx) parseDocument(file *zip.File) error {
	zf, err := file.Open()
	if err != nil {
		return err
	}
	defer zf.Close()

	f.Document.XMLW = XMLNS_W
	f.Document.XMLR = XMLNS_R
	f.Document.XMLWP = XMLNS_WP
	f.Document.XMLWP14 = XMLNS_WP14
	f.Document.XMLName.Space = XMLNS_W
	f.Document.XMLName.Local = "document"
	err = xml.NewDecoder(zf).Decode(&f.Document)
	if err != nil {
		return err
	}
	return nil
}

// parseDocRelation processes one of the relevant files, the one with the relationships
func (f *Docx) parseDocRelation(file *zip.File) error {
	zf, err := file.Open()
	if err != nil {
		return err
	}
	defer zf.Close()

	f.DocRelation.Xmlns = XMLNS_R
	//TODO: find last rId & imageId
	return xml.NewDecoder(zf).Decode(&f.DocRelation)
}

func (f *Docx) checkAndParseMedia(file *zip.File) error {
	if !strings.HasPrefix(file.Name, MEDIA_FOLDER) {
		return nil
	}
	name := file.Name[len(MEDIA_FOLDER):]
	zf, err := file.Open()
	if err != nil {
		return err
	}
	data, err := io.ReadAll(zf)
	if err != nil {
		return err
	}
	f.mediaNameIdx[name] = len(f.media)
	f.media = append(f.media, Media{Name: name, Data: data})
	return zf.Close()
}
