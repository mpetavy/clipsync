package main

type Bookmarks struct {
	Children []struct {
		Children []struct {
			Children []struct {
				DateAdded int64  `json:"dateAdded"`
				Id        string `json:"id"`
				Index     int    `json:"index"`
				ParentId  string `json:"parentId"`
				Syncing   bool   `json:"syncing"`
				Title     string `json:"title"`
				Url       string `json:"url,omitempty"`
				Children  []struct {
					DateAdded int64  `json:"dateAdded"`
					Id        string `json:"id"`
					Index     int    `json:"index"`
					ParentId  string `json:"parentId"`
					Syncing   bool   `json:"syncing"`
					Title     string `json:"title"`
					Url       string `json:"url,omitempty"`
					Children  []struct {
						DateAdded         int64         `json:"dateAdded"`
						Id                string        `json:"id"`
						Index             int           `json:"index"`
						ParentId          string        `json:"parentId"`
						Syncing           bool          `json:"syncing"`
						Title             string        `json:"title"`
						Url               string        `json:"url,omitempty"`
						Children          []interface{} `json:"children,omitempty"`
						DateGroupModified int64         `json:"dateGroupModified,omitempty"`
					} `json:"children,omitempty"`
				} `json:"children,omitempty"`
				DateGroupModified int64 `json:"dateGroupModified,omitempty"`
			} `json:"children,omitempty"`
			DateAdded         int64  `json:"dateAdded"`
			Id                string `json:"id"`
			Index             int    `json:"index"`
			ParentId          string `json:"parentId"`
			Syncing           bool   `json:"syncing"`
			Title             string `json:"title"`
			Url               string `json:"url,omitempty"`
			DateGroupModified int64  `json:"dateGroupModified,omitempty"`
		} `json:"children"`
		DateAdded  int64  `json:"dateAdded"`
		FolderType string `json:"folderType"`
		Id         string `json:"id"`
		Index      int    `json:"index"`
		ParentId   string `json:"parentId"`
		Syncing    bool   `json:"syncing"`
		Title      string `json:"title"`
	} `json:"children"`
	DateAdded int64  `json:"dateAdded"`
	Id        string `json:"id"`
	Syncing   bool   `json:"syncing"`
	Title     string `json:"title"`
}
