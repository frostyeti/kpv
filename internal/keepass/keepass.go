package keepass

// TODO: consider using a second graph that uses the internal Group and Entry
// structs to build the tree. This would allow for more efficient lookups
// and would allow for more efficient tree traversals.
import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
)

const TEST = "TEST"

func init() {
}

var (
	ErrNilKdbx  = errors.New("kdbx is nil")
	ErrNoSecret = errors.New("no secret provided")
)

type KdbxOptions struct {
	Path                string
	Secret              *string
	SecretFileData      []byte
	Create              bool
	CreateDir           bool
	UseCommonDelimiters bool
	Delimiter           *string
}

type Kdbx struct {
	db        *gokeepasslib.Database
	options   KdbxOptions
	open      bool
	delimiter string
	rootGroup *Group
}

type pathQuery []string

func New(options KdbxOptions) *Kdbx {
	return &Kdbx{
		options:   options,
		open:      false,
		delimiter: "/",
	}
}

func Create(options KdbxOptions) (*Kdbx, error) {
	kdbx := New(options)
	return kdbx, kdbx.Create()
}

func Open(options KdbxOptions) (*Kdbx, error) {
	kdbx := New(options)
	return kdbx, kdbx.Open()
}

func (kdbx *Kdbx) Create() error {
	if kdbx == nil {
		return ErrNilKdbx
	}

	var creds *gokeepasslib.DBCredentials
	if kdbx.options.Secret != nil && kdbx.options.SecretFileData != nil {
		c, err := gokeepasslib.NewPasswordAndKeyDataCredentials(*kdbx.options.Secret, kdbx.options.SecretFileData)
		if err != nil {
			return err
		}

		creds = c
	} else if kdbx.options.Secret != nil {
		c := gokeepasslib.NewPasswordCredentials(*kdbx.options.Secret)

		creds = c
	} else if kdbx.options.SecretFileData != nil {
		c, err := gokeepasslib.NewKeyDataCredentials(kdbx.options.SecretFileData)
		if err != nil {
			return err
		}

		creds = c
	} else {
		return ErrNoSecret
	}

	if _, err := os.Stat(kdbx.options.Path); err != nil {
		if !os.IsNotExist(err) {
			if os.IsExist(err) {
				return errors.New("file already exists")
			} else {
				return err
			}
		}
	}

	db := gokeepasslib.NewDatabase(
		gokeepasslib.WithDatabaseKDBXVersion4(),
	)

	rg := rootGroupFromPath(kdbx.options.Path)
	db.Content.Meta.DatabaseName = rg
	db.Content.Root.Groups[0].Name = rg
	db.Credentials = creds

	dir := filepath.Dir(kdbx.options.Path)

	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			if !kdbx.options.CreateDir {
				return errors.New("directory does not exist and createDir option is false")
			}
		} else {
			return err
		}
	}

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(kdbx.options.Path)
	if err != nil {
		return err
	}

	err = db.LockProtectedEntries()
	if err != nil {
		return err
	}
	enc := gokeepasslib.NewEncoder(file)
	if err := enc.Encode(db); err != nil {
		file.Close()
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	err = db.UnlockProtectedEntries()
	if err != nil {
		return err
	}

	kdbx.open = true
	kdbx.db = db
	return nil

}

func (kdbx *Kdbx) Open() error {
	if kdbx == nil {
		return errors.New("kdbx is nil")
	}

	if kdbx.open {
		return nil
	}

	var creds *gokeepasslib.DBCredentials
	db := gokeepasslib.NewDatabase()
	if kdbx.options.Secret != nil && kdbx.options.SecretFileData != nil {
		c, err := gokeepasslib.NewPasswordAndKeyDataCredentials(*kdbx.options.Secret, kdbx.options.SecretFileData)
		if err != nil {
			return err
		}

		creds = c
	} else if kdbx.options.Secret != nil {
		c := gokeepasslib.NewPasswordCredentials(*kdbx.options.Secret)

		creds = c
	} else if kdbx.options.SecretFileData != nil {
		c, err := gokeepasslib.NewKeyDataCredentials(kdbx.options.SecretFileData)
		if err != nil {
			return err
		}

		creds = c
	} else {
		return ErrNoSecret
	}

	exists := true

	if _, err := os.Stat(kdbx.options.Path); err != nil {
		if os.IsNotExist(err) {
			exists = false
		} else {
			return err
		}
	}

	dirname := filepath.Dir(kdbx.options.Path)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if !kdbx.options.CreateDir {
			return errors.New("directory does not exist and createDir option is false")
		}

		err := os.MkdirAll(dirname, 0755)
		if err != nil {
			return err
		}
	}

	if !exists {

		if !kdbx.options.Create {
			return errors.New("file does not exist and create option is false")
		}

		db = gokeepasslib.NewDatabase(
			gokeepasslib.WithDatabaseKDBXVersion4(),
		)

		rg := rootGroupFromPath(kdbx.options.Path)
		db.Content.Meta.DatabaseName = rg
		db.Content.Root.Groups[0].Name = rg
		db.Credentials = creds

		file, err := os.Create(kdbx.options.Path)
		if err != nil {
			return err
		}

		err = db.LockProtectedEntries()
		if err != nil {
			return err
		}
		enc := gokeepasslib.NewEncoder(file)
		if err := enc.Encode(db); err != nil {
			file.Close()
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}

		err = db.UnlockProtectedEntries()
		if err != nil {
			return err
		}

		kdbx.open = true
		kdbx.db = db
		return nil
	}

	db.Credentials = creds
	file, err := os.OpenFile(kdbx.options.Path, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	dec := gokeepasslib.NewDecoder(file)
	if err := dec.Decode(db); err != nil {
		return err
	}

	err = db.UnlockProtectedEntries()
	if err != nil {
		return err
	}

	kdbx.open = true
	kdbx.db = db
	return nil
}

func (kdbx *Kdbx) Save() error {
	if kdbx.db.Credentials == nil {

		if kdbx.options.Secret != nil && kdbx.options.SecretFileData != nil {
			c, err := gokeepasslib.NewPasswordAndKeyDataCredentials(*kdbx.options.Secret, kdbx.options.SecretFileData)
			if err != nil {
				return err
			}
			kdbx.db.Credentials = c
		} else if kdbx.options.Secret != nil {
			c := gokeepasslib.NewPasswordCredentials(*kdbx.options.Secret)
			kdbx.db.Credentials = c
		} else if kdbx.options.SecretFileData != nil {
			c, err := gokeepasslib.NewKeyDataCredentials(kdbx.options.SecretFileData)
			if err != nil {
				return err
			}
			kdbx.db.Credentials = c
		} else {
			return ErrNoSecret
		}
	}

	// ensure the root group is correctly set before saving
	if kdbx.rootGroup != nil {
		kdbx.db.Content.Root.Groups[0] = *kdbx.rootGroup.Group
	}

	err := kdbx.db.LockProtectedEntries()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(kdbx.options.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	defer kdbx.db.UnlockProtectedEntries()

	enc := gokeepasslib.NewEncoder(file)
	if err := enc.Encode(kdbx.db); err != nil {
		return err
	}

	return nil
}

func (kdbx *Kdbx) GetBinaries() *gokeepasslib.Binaries {
	if kdbx == nil || kdbx.db == nil {
		return nil
	}

	if kdbx.db.Header.IsKdbx4() {
		return &kdbx.db.Content.InnerHeader.Binaries
	}

	return &kdbx.db.Content.Meta.Binaries
}

func (kdbx *Kdbx) SaveAs(path string) error {
	err := kdbx.db.LockProtectedEntries()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	defer kdbx.db.UnlockProtectedEntries()

	enc := gokeepasslib.NewEncoder(file)
	if err := enc.Encode(kdbx.db); err != nil {
		return err
	}

	return nil
}

func (kdbx *Kdbx) IsOpen() bool {
	if kdbx == nil {
		return false
	}
	return kdbx.open
}

func rootGroupFromPath(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	if ext != "" {
		base = base[:len(base)-len(ext)]
	}
	return base
}

func (kdbx *Kdbx) Root() *Group {
	if kdbx == nil {
		return nil
	}

	if kdbx.rootGroup != nil {
		return kdbx.rootGroup
	}

	if kdbx.db == nil {
		return nil
	}

	root := kdbx.db.Content.Root
	if len(root.Groups) == 0 {
		root.Groups = make([]gokeepasslib.Group, 1)
		group := gokeepasslib.NewGroup()
		group.Name = rootGroupFromPath(kdbx.options.Path)
		root.Groups[0] = group
	}

	group := kdbx.db.Content.Root.Groups[0]
	kdbx.rootGroup = &Group{
		Group:  &group,
		parent: nil,
	}

	return kdbx.rootGroup
}

func (kdbx *Kdbx) FindGroup(path string) *Group {
	query := kdbx.splitPath(path)
	return kdbx.findGroup(query)
}

func (kdbx *Kdbx) FindEntry(path string) *Entry {
	query := kdbx.splitPath(path)
	if len(query) == 1 {
		name := query[0]
		group := kdbx.Root()
		if group == nil {
			return nil
		}

		if len(group.Entries) == 0 {
			return nil
		}

		for _, entry := range group.Entries {
			title := entry.GetContent(KP_TITLE)
			if strings.EqualFold(title, name) {

				return &Entry{
					Entry:  &entry,
					parent: group,
				}
			}

			index := entry.GetIndex(KP_PATH)
			if index == -1 {
				continue
			}

			altPath := entry.GetContent(KP_PATH)
			if strings.EqualFold(altPath, name) {
				return &Entry{
					&entry, group,
				}
			}
		}

		return nil
	}

	lastIndex := len(query) - 1
	name := query[lastIndex]
	group := kdbx.findGroup(query[:lastIndex])
	if group == nil {
		return nil
	}

	for _, entry := range group.Entries {
		title := entry.GetContent(KP_TITLE)
		if strings.EqualFold(title, name) {
			return &Entry{
				Entry:  &entry,
				parent: group,
			}
		}

		index := entry.GetIndex(KP_PATH)
		if index == -1 {
			continue
		}

		altPath := entry.GetContent(KP_PATH)
		if strings.EqualFold(altPath, name) {
			return &Entry{
				Entry:  &entry,
				parent: group,
			}
		}
	}

	return nil
}

func (kdbx *Kdbx) UpsertEntry(path string, cb func(entry *Entry)) *Entry {
	query := kdbx.splitPath(path)
	if len(query) == 1 {
		name := query[0]
		group := kdbx.Root()

		if len(group.Entries) > 0 {
			for i := range group.Entries {
				title := group.Entries[i].GetTitle()
				if strings.EqualFold(title, name) {
					e := &Entry{
						&group.Entries[i], group,
					}
					cb(e)

					return e
				}
			}
		}

		entry := NewEntry()
		entry.SetTitle(name)
		cb(entry)
		group.Entries = append(group.Entries, *entry.Entry)
		return entry
	}

	lastIndex := len(query) - 1
	name := query[lastIndex]
	groupQuery := query[:lastIndex]
	group := kdbx.Root()

	groups := group.Groups
	prevIndex := 0
	for _, seg := range groupQuery {
		found := false
		for i, nextGroup := range groups {
			if strings.EqualFold(nextGroup.Name, seg) {
				prevIndex = i

				groups = nextGroup.Groups
				found = true
				break
			}
		}

		if found {
			continue
		}

		if !found {
			ng := NewGroup()
			ng.Name = seg
			group.Groups = append(group.Groups, *ng.Group)
			groups[prevIndex] = *ng.Group
			groups = group.Groups
			prevIndex = 0
			group = &Group{
				ng.Group, group,
			}
		}
	}

	for i := range group.Entries {
		title := group.Entries[i].GetTitle()
		if strings.EqualFold(title, name) {
			e := &Entry{
				&group.Entries[i],
				group,
			}
			cb(e)
			groups[prevIndex] = *group.Group
			return e
		}
	}

	entry := NewEntry()
	entry.SetTitle(name)
	cb(entry)
	group.Entries = append(group.Entries, *entry.Entry)
	return entry
}

func (kdbx *Kdbx) findGroup(query pathQuery) *Group {
	group := kdbx.Root()

	groups := group.Groups
	for _, seg := range query {
		found := false
		for _, nextGroup := range groups {
			if strings.EqualFold(nextGroup.Name, seg) {
				group = &Group{
					&nextGroup, group,
				}
				groups = group.Groups
				found = true
				break
			}
		}

		if !found {
			return nil
		}
	}

	return group
}

func (kdbx *Kdbx) splitPath(path string) pathQuery {
	if kdbx.options.Delimiter != nil && *kdbx.options.Delimiter != "" {
		return strings.Split(*kdbx.options.Delimiter, path)
	}

	return splitAny(path, "\\/.:")

}

func splitAny(s, sep string) []string {
	set := make([]string, 0)
	sb := strings.Builder{}
	for _, r := range s {
		if strings.ContainsRune(sep, r) {
			set = append(set, sb.String())
			sb.Reset()
			continue
		}

		sb.WriteRune(r)
	}

	if sb.Len() > 0 {
		set = append(set, sb.String())
	}

	return set
}
