package main

import (
	"errors"
	"io"
	"os"

	pbar "github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	// Открываем исходный файл
	srcFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Получаем информацию о файле
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()

	// Проверяем, что файл имеет известный размер
	if fileSize == -1 {
		return ErrUnsupportedFile
	}

	// Проверяем, что offset не превышает размер файла
	if offset > fileSize {
		return ErrOffsetExceedsFileSize
	}

	// Вычисляем количество байт для копирования
	bytesToCopy := fileSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	// Создаем целевой файл
	dstFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Перемещаемся к offset в исходном файле
	_, err = srcFile.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// Создаем прогресс-бар
	bar := pbar.StartNew(int(bytesToCopy))
	bar.SetTemplateString(`{{counters . }} {{bar . }} {{percent . }}`)
	bar.Set(pbar.Bytes, true)

	// Создаем writer для прогресс-бара
	writer := bar.NewProxyWriter(dstFile)

	// Копируем данные с прогресс-баром
	_, err = io.CopyN(writer, srcFile, bytesToCopy)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	bar.Finish()

	return nil
}
