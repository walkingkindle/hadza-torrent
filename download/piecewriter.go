package download

import (
	"os"
	"path/filepath"

	"torrent-client-go/types"
)

type PieceWriter interface {
	WritePiece(pieceIndex int, data []byte) error
}

type TorrentWriter struct {
	files       []fileentry
	pieceLength int
}

type fileentry struct {
	file   *os.File
	path   string
	offset int64
	length int64
}

func NewTorrentWriter(root string, torrent types.TorrentFile) (TorrentWriter, error) {
	files := make([]fileentry, 0, len(torrent.Files))

	var offset int64

	for _, torrentFile := range torrent.Files {
		path := filepath.Join(
			append([]string{root}, torrentFile.Path...)...,
		)

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return TorrentWriter{}, err
		}

		f, err := os.OpenFile(
			path,
			os.O_CREATE|os.O_RDWR,
			0o644,
		)
		if err != nil {
			return TorrentWriter{}, err
		}

		if err := f.Truncate(torrentFile.Length); err != nil {
			f.Close()
			return TorrentWriter{}, err
		}

		files = append(files, fileentry{
			file:   f,
			path:   path,
			offset: offset,
			length: torrentFile.Length,
		})

		offset += torrentFile.Length
	}

	return TorrentWriter{
		files:       files,
		pieceLength: torrent.PieceLength,
	}, nil
}

func (w *TorrentWriter) WritePiece(
	pieceIndex int,
	data []byte,
) error {
	pieceOffset := int64(pieceIndex * w.pieceLength)
	pieceEnd := pieceOffset + int64(len(data))

	for _, file := range w.files {
		fileStart := file.offset
		fileEnd := file.offset + file.length

		// No overlap.
		if pieceEnd <= fileStart || pieceOffset >= fileEnd {
			continue
		}

		writeStart := max(pieceOffset, fileStart)
		writeEnd := min(pieceEnd, fileEnd)

		// Offset inside the piece.
		dataStart := writeStart - pieceOffset
		dataEnd := writeEnd - pieceOffset

		// Offset inside the destination file.
		fileOffset := writeStart - fileStart

		f, err := os.OpenFile(
			file.path,
			os.O_WRONLY,
			0,
		)
		if err != nil {
			return err
		}

		_, err = f.WriteAt(
			data[int(dataStart):int(dataEnd)],
			fileOffset,
		)

		closeErr := f.Close()

		if err != nil {
			return err
		}

		if closeErr != nil {
			return closeErr
		}
	}

	return nil
}
