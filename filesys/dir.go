package filesys

import (
	"log"
	"os"

	"golang.org/x/net/context"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

type Dir struct {
	Node
	Files       *[]*File
	Directories *[]*Dir
}

func (dir *Dir) Attr(ctx context.Context, attr *fuse.Attr) error {
	log.Println("Attributes for directory", dir.Name)
	attr.Inode = dir.Inode
	attr.Mode = os.ModeDir | 0444
	return nil
}

func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Lookup for ", name)
	if dir.Files != nil {
		for _, file := range *dir.Files {
			if file.Name == name {
				log.Println("Found match for directory lookup with size", len(file.Data))
				return file, nil
			}
		}
	}
	if dir.Directories != nil {
		for _, directory := range *dir.Directories {
			if directory.Name == name {
				log.Println("Found match for directory lookup")
				return directory, nil
			}
		}
	}
	return nil, fuse.ENOENT
}

func (dir *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	log.Println("Mkdir request for name", req.Name)
	newDir := &Dir{Node: Node{Name: req.Name, Inode: NewInode()}}
	directories := []*Dir{newDir}
	if dir.Directories != nil {
		directories = append(*dir.Directories, directories...)
	}
	dir.Directories = &directories
	return newDir, nil

}

func (dir *Dir) ReadDir(ctx context.Context, name string) (fs.Node, error) {
	if dir.Files != nil {
		for _, file := range *dir.Files {
			if file.Name == name {
				return file, nil
			}
		}
	}
	if dir.Directories != nil {
		for _, directory := range *dir.Directories {
			if directory.Name == name {
				return directory, nil
			}
		}
	}
	return nil, fuse.ENOENT
}

func (dir *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Create request for name", req.Name)
	newFile := &File{Node: Node{Name: req.Name, Inode: NewInode()}}
	files := []*File{newFile}
	if dir.Files != nil {
		files = append(files, *dir.Files...)
	}
	dir.Files = &files
	return newFile, newFile, nil
}

func (dir *Dir) Remove(ctx context.Context, req *fuse.RemoveRequest) error {
	log.Println("Remove request for ", req.Name)
	if req.Dir && dir.Directories != nil {
		newDirs := []*Dir{}
		for _, directory := range *dir.Directories {
			if directory.Name != req.Name {
				newDirs = append(newDirs, directory)
			}
		}
		dir.Directories = &newDirs
		return nil
	} else if !req.Dir && *dir.Files != nil {
		newFiles := []*File{}
		for _, file := range *dir.Files {
			if file.Name != req.Name {
				newFiles = append(newFiles, file)
			}
		}
		dir.Files = &newFiles
		return nil
	}
	return fuse.ENOENT
}

func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Reading all dirs")
	var children []fuse.Dirent
	if dir.Files != nil {
		for _, file := range *dir.Files {
			children = append(children, fuse.Dirent{Inode: file.Inode, Type: fuse.DT_File, Name: file.Name})
		}
	}
	if dir.Directories != nil {
		for _, directory := range *dir.Directories {
			children = append(children, fuse.Dirent{Inode: directory.Inode, Type: fuse.DT_Dir, Name: directory.Name})
		}
		log.Println(len(children), " children for dir", dir.Name)
	}
	return children, nil
}