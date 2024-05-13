package test

type Artist struct {
	Id      string
	Name    string
	Picture string
}

type Album struct {
	Id       string
	Name     string
	CoverArt string
	ArtistId string
}

type Track struct {
	Id                string
	Number            int
	Name              string
	CoverArt          string
	Duration          int
	BestQualityFile   string
	MobileQualityFile string
	AlbumId           string
	ArtistId          string
	AlbumName         string
	ArtistName        string
}

type GetArtists struct {
	Artists []Artist
}

type GetArtistById struct {
	Artist
}

type GetArtistAlbumsById struct {
	Albums []Album
}

type GetAlbums struct {
	Albums []Album
}

type GetAlbumById struct {
	Album
}

type GetAlbumTracksById struct {
	Tracks []Track
}

type GetTracks struct {
	Tracks []Track
}

type GetTrackById struct {
	Track
	ArtistName int
}

type GetSync struct {
	IsSyncing bool
}

type Test struct {
	Name string
}

type Test2 struct {
	Test  *Test `json:"test"`
	Test2 int `json:"test2,omitempty"`
}

// TODO(patrik): This is not working

// type TestStruct struct {
// 	Field1 string
// 	Wooh string
// }
//
// type TestStruct3 struct {
// 	TestStruct
// }
//
// type TestStruct2 struct {
// 	TestStruct3
//
// 	Field2, Hello int
// }
