package pile

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/fletaio/fleta_v1/common/binutil"
	"github.com/fletaio/fleta_v1/common/hash"
)

// Pile proivdes a part of stack like store
type Pile struct {
	sync.Mutex
	file        *os.File
	HeadHeight  uint32
	BeginHeight uint32
	GenHash     hash.Hash256
}

// NewPile returns a Pile
func NewPile(path string, GenHash hash.Hash256, BaseHeight uint32) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	if fi, err := file.Stat(); err != nil {
		file.Close()
		return nil, err
	} else if fi.Size() < ChunkHeaderSize {
		if err := file.Truncate(ChunkHeaderSize); err != nil {
			file.Close()
			return nil, err
		}
		if _, err := file.Seek(0, 0); err != nil {
			file.Close()
			return nil, err
		}
		meta := make([]byte, ChunkMetaSize)
		if BaseHeight%ChunkUnit != 0 {
			file.Close()
			return nil, ErrInvalidChunkBeginHeight
		}
		copy(meta, binutil.LittleEndian.Uint32ToBytes(BaseHeight))                //HeadHeight (0, 4)
		copy(meta[4:], binutil.LittleEndian.Uint32ToBytes(BaseHeight))            //HeadHeightCheckA (4, 8)
		copy(meta[8:], binutil.LittleEndian.Uint32ToBytes(BaseHeight))            //HeadHeightCheckB (8, 12)
		copy(meta[12:], binutil.LittleEndian.Uint32ToBytes(BaseHeight))           //BeginHeight (12, 16)
		copy(meta[16:], binutil.LittleEndian.Uint32ToBytes(BaseHeight+ChunkUnit)) //EndHeight (16, 20)
		copy(meta[20:], GenHash[:])                                               //GenesisHash (20, 52)
		if _, err := file.Write(meta); err != nil {
			file.Close()
			return nil, err
		}
	}

	p := &Pile{
		file:        file,
		HeadHeight:  BaseHeight,
		BeginHeight: BaseHeight,
		GenHash:     GenHash,
	}
	return p, nil
}

// LoadPile loads a pile from the file
func LoadPile(path string) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return nil, err
	}
	meta := make([]byte, ChunkMetaSize)
	if _, err := file.Read(meta); err != nil {
		file.Close()
		return nil, err
	}
	HeadHeight := binutil.LittleEndian.Uint32(meta)
	HeadHeightCheckA := binutil.LittleEndian.Uint32(meta[4:])
	HeadHeightCheckB := binutil.LittleEndian.Uint32(meta[8:])
	BeginHeight := binutil.LittleEndian.Uint32(meta[12:])
	EndHeight := binutil.LittleEndian.Uint32(meta[16:])
	var GenHash hash.Hash256
	copy(GenHash[:], meta[20:])
	if BeginHeight%ChunkUnit != 0 {
		file.Close()
		return nil, ErrInvalidChunkBeginHeight
	}
	if BeginHeight+ChunkUnit != EndHeight {
		file.Close()
		return nil, ErrInvalidChunkEndHeight
	}
	if HeadHeight != HeadHeightCheckA || HeadHeightCheckA != HeadHeightCheckB { // crashed when file.Sync()
		if HeadHeightCheckA == HeadHeightCheckB { //crashed at HeadHeight
			if _, err := file.Seek(0, 0); err != nil {
				return nil, err
			}
			if _, err := file.Write(binutil.LittleEndian.Uint32ToBytes(HeadHeightCheckA)); err != nil {
				return nil, err
			}
			if err := file.Sync(); err != nil {
				return nil, err
			}
			HeadHeight = HeadHeightCheckA
		} else if HeadHeight == HeadHeightCheckB+1 { //crashed at HeadHeightCheckA
			if _, err := file.Seek(4, 0); err != nil {
				return nil, err
			}
			if _, err := file.Write(binutil.LittleEndian.Uint32ToBytes(HeadHeight)); err != nil {
				return nil, err
			}
			if _, err := file.Write(binutil.LittleEndian.Uint32ToBytes(HeadHeight)); err != nil {
				return nil, err
			}
			if err := file.Sync(); err != nil {
				return nil, err
			}
			HeadHeightCheckA = HeadHeight
			HeadHeightCheckB = HeadHeight
		} else if HeadHeight == HeadHeightCheckA { //crashed at HeadHeightCheckB
			if _, err := file.Seek(8, 0); err != nil {
				return nil, err
			}
			if _, err := file.Write(binutil.LittleEndian.Uint32ToBytes(HeadHeight)); err != nil {
				return nil, err
			}
			if err := file.Sync(); err != nil {
				return nil, err
			}
			HeadHeightCheckB = HeadHeight
		} else {
			log.Println("PileDB height crashed", HeadHeight, HeadHeightCheckA, HeadHeightCheckB)
			return nil, ErrHeightCrashed
		}
	}
	if true {
		Offset := ChunkHeaderSize
		FromHeight := HeadHeight - BeginHeight
		if FromHeight > 0 {
			if _, err := file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
				file.Close()
				return nil, err
			}
			bs := make([]byte, 8)
			if _, err := file.Read(bs); err != nil {
				file.Close()
				return nil, err
			}
			Offset = int64(binutil.LittleEndian.Uint64(bs))
		}
		if fi, err := file.Stat(); err != nil {
			file.Close()
			return nil, err
		} else if fi.Size() < Offset {
			return nil, ErrInvalidFileSize
		}
	}
	p := &Pile{
		file:        file,
		HeadHeight:  HeadHeight,
		BeginHeight: BeginHeight,
		GenHash:     GenHash,
	}
	return p, nil
}

// Close closes a pile
func (p *Pile) Close() {
	p.Lock()
	defer p.Unlock()

	if p.file != nil {
		p.file.Sync()
		p.file.Close()
		p.file = nil
	}
}

// AppendData pushes data to the top of the pile
func (p *Pile) AppendData(Sync bool, Height uint32, DataHash hash.Hash256, Datas [][]byte) error {
	p.Lock()
	defer p.Unlock()

	if Height != p.HeadHeight+1 {
		return ErrInvalidAppendHeight
	}

	FromHeight := p.HeadHeight - p.BeginHeight

	//get offset
	Offset := ChunkHeaderSize
	if FromHeight > 0 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
			return err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return err
		}
		Offset = int64(binutil.LittleEndian.Uint64(bs))
	}

	// write data
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return err
	}
	totalLen := int64(32 + 1 + 4*len(Datas))
	if _, err := p.file.Write(DataHash[:]); err != nil {
		return err
	}
	if _, err := p.file.Write([]byte{uint8(len(Datas))}); err != nil {
		return err
	}
	zdatas := make([][]byte, 0, len(Datas))
	for _, v := range Datas {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		if _, err := zw.Write(v); err != nil {
			return err
		}
		zw.Flush()
		zw.Close()
		zd := buffer.Bytes()
		zdatas = append(zdatas, zd)

		if _, err := p.file.Write(binutil.LittleEndian.Uint32ToBytes(uint32(len(zd)))); err != nil {
			return err
		}
	}
	for _, zd := range zdatas {
		if _, err := p.file.Write(zd); err != nil {
			return err
		}
		totalLen += int64(len(zd))
	}

	// update offset
	if _, err := p.file.Seek(ChunkMetaSize+int64(FromHeight)*8, 0); err != nil {
		return err
	}
	if _, err := p.file.Write(binutil.LittleEndian.Uint64ToBytes(uint64(Offset + totalLen))); err != nil {
		return err
	}

	// update head height
	if _, err := p.file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := p.file.Write(binutil.LittleEndian.Uint32ToBytes(p.HeadHeight + 1)); err != nil {
		return err
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return err
		}
	}

	// update head height check A
	if _, err := p.file.Seek(4, 0); err != nil {
		return err
	}
	if _, err := p.file.Write(binutil.LittleEndian.Uint32ToBytes(p.HeadHeight + 1)); err != nil {
		return err
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return err
		}
	}

	// update head height check B
	if _, err := p.file.Seek(8, 0); err != nil {
		return err
	}
	if _, err := p.file.Write(binutil.LittleEndian.Uint32ToBytes(p.HeadHeight + 1)); err != nil {
		return err
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return err
		}
	}
	p.HeadHeight++
	return nil
}

// GetHash returns a hash value of the height
func (p *Pile) GetHash(Height uint32) (hash.Hash256, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return hash.Hash256{}, ErrInvalidHeight
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return hash.Hash256{}, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return hash.Hash256{}, err
		}
		Offset = int64(binutil.LittleEndian.Uint64(bs))
	}
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return hash.Hash256{}, err
	}
	value := make([]byte, 32)
	if _, err := p.file.Read(value); err != nil {
		return hash.Hash256{}, err
	}
	var h hash.Hash256
	copy(h[:], value)
	return h, nil
}

// GetData returns a data at the index of the height
func (p *Pile) GetData(Height uint32, index int) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return nil, ErrInvalidHeight
	}
	if Height > p.HeadHeight {
		return nil, ErrInvalidHeight
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return nil, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, err
		}
		Offset = int64(binutil.LittleEndian.Uint64(bs))
	}
	if _, err := p.file.Seek(Offset+32, 0); err != nil {
		return nil, err
	}
	lbs := make([]byte, 1)
	if _, err := p.file.Read(lbs); err != nil {
		return nil, err
	}
	if index >= int(lbs[0]) {
		return nil, ErrInvalidDataIndex
	}
	zlbs := make([]byte, 4*lbs[0])
	if _, err := p.file.Read(zlbs); err != nil {
		return nil, err
	}
	zofs := Offset + 32 + 1 + int64(4*lbs[0])
	for i := 0; i < index; i++ {
		zofs += int64(binutil.LittleEndian.Uint32(zlbs[4*i:]))
	}
	if _, err := p.file.Seek(zofs, 0); err != nil {
		return nil, err
	}

	zsize := binutil.LittleEndian.Uint32(zlbs[4*index:])
	zd := make([]byte, zsize)
	if _, err := p.file.Read(zd); err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(zd))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// GetDatas returns datas of the height between from and from + count
func (p *Pile) GetDatas(Height uint32, from int, count int) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return nil, ErrInvalidHeight
	}
	if Height > p.HeadHeight {
		return nil, ErrInvalidHeight
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return nil, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, err
		}
		Offset = int64(binutil.LittleEndian.Uint64(bs))
	}
	if _, err := p.file.Seek(Offset+32, 0); err != nil {
		return nil, err
	}
	lbs := make([]byte, 1)
	if _, err := p.file.Read(lbs); err != nil {
		return nil, err
	}
	if from+count > int(lbs[0]) {
		return nil, ErrInvalidDataIndex
	}
	zlbs := make([]byte, 4*lbs[0])
	if _, err := p.file.Read(zlbs); err != nil {
		return nil, err
	}
	zofs := Offset + 32 + 1 + int64(4*lbs[0])
	for i := 0; i < from; i++ {
		zofs += int64(binutil.LittleEndian.Uint32(zlbs[4*i:]))
	}
	if _, err := p.file.Seek(zofs, 0); err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	for i := 0; i < count; i++ {
		zsize := binutil.LittleEndian.Uint32(zlbs[4*(from+i):])
		zd := make([]byte, zsize)
		if _, err := p.file.Read(zd); err != nil {
			return nil, err
		}
		zr, err := gzip.NewReader(bytes.NewReader(zd))
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(zr)
		if err != nil {
			return nil, err
		}
		buffer.Write(data)
	}
	return buffer.Bytes(), nil
}
