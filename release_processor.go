package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

var inputReleases = flag.String("input_releases", "/Users/Shared/Documents/release/mbdump/release", "Input releases json file from Musicbrainz.")
var outputArtists = flag.String("output_artists", "/Users/Shared/Documents/soundstory/data/artists.json", "Output artists json file.")

type Release struct {
	Title   string
	Date    string
	Artwork string
}

type Artist struct {
	Name     string
	Releases []Release
}

func Insert(releases []Release, release Release) []Release {
	i := sort.Search(len(releases), func(x int) bool {
		return releases[x].Date >= release.Date
	})
	releases = append(releases, Release{})
	copy(releases[i+1:], releases[i:])
	releases[i] = release
	return releases
}

func queryCoverArt(mbid string) string {
	fmt.Println(fmt.Sprintf("https://coverartarchive.org/release/%s", mbid))
	resp, err := http.Get(fmt.Sprintf("https://coverartarchive.org/release/%s", mbid))
	if err != nil {
		fmt.Printf("coverart query %v\n", err)
		os.Exit(1)
	}
	if resp.StatusCode == 404 {
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("coverart response %v\n", err)
		os.Exit(1)
	}
	data := make(map[string]json.RawMessage)
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Printf("unmarshal coverart %v\n", err)
		os.Exit(1)
	}
	var images []json.RawMessage
	if err := json.Unmarshal(data["images"], &images); err != nil {
		fmt.Printf("unmarshal images: %v\n", err)
		os.Exit(1)
	}
	if len(images) == 0 {
		return ""
	}
	var image map[string]json.RawMessage
	if err := json.Unmarshal(images[0], &image); err != nil {
		fmt.Printf("unmarshal image %v\n", err)
		os.Exit(1)
	}
	var imageURL string
	if err := json.Unmarshal(image["image"], &imageURL); err != nil {
		fmt.Printf("unmarshal image URL %v\n", err)
		os.Exit(1)
	}
	return imageURL
}

func parseRecord(decoder *json.Decoder) ([]string, Release) {
	data := make(map[string]json.RawMessage)
	decoder.Decode(&data)

	var artistMeta []json.RawMessage
	if err := json.Unmarshal(data["artist-credit"], &artistMeta); err != nil {
		fmt.Printf("unmarshal artist %v\n", err)
		os.Exit(1)
	}
	var artists []string
	for _, meta := range artistMeta {
		metamap := make(map[string]json.RawMessage)
		if err := json.Unmarshal(meta, &metamap); err != nil {
			fmt.Printf("unmarshal artist meta %v\n", err)
			os.Exit(1)
		}
		var artist string
		if err := json.Unmarshal(metamap["name"], &artist); err != nil {
			fmt.Printf("unmarshal artist name %v\n", err)
			os.Exit(1)
		}
		artists = append(artists, artist)
	}

	var title string
	if err := json.Unmarshal(data["title"], &title); err != nil {
		fmt.Printf("unmarshal title %v\n", err)
		os.Exit(1)
	}

	var date string
	if _, ok := data["date"]; ok {
		if err := json.Unmarshal(data["date"], &date); err != nil {
			fmt.Printf("unmarshal date %v: %v\n", data["date"], err)
			os.Exit(1)
		}
	}

	var mbid string
	if err := json.Unmarshal(data["id"], &mbid); err != nil {
		fmt.Printf("unmarshal mbid %v\n", err)
		os.Exit(1)
	}

	// read the cover art
	// art := queryCoverArt(mbid)

	return artists, Release{
		Title: title,
		Date:  date,
		// Artwork: art,
	}
}

func main() {
	// read record by record
	jsonFile, err := os.Open(*inputReleases)
	if err != nil {
		fmt.Printf("fileopen: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	// parse the data
	decoder := json.NewDecoder(jsonFile)
	// artists := make(map[string]Artist)
	var numSongs int
	for decoder.More() {
		numSongs++
		_, release := parseRecord(decoder)
		fmt.Printf("Processing song %v: %v\n", numSongs, release.Title)

		// for _, name := range artistNames {
		// 	artist, ok := artists[name]
		// 	if !ok {
		// 		artist = Artist{
		// 			Name: name,
		// 		}
		// 		artists[artist.Name] = artist
		// 	}

		// 	artist.Releases = Insert(artist.Releases, release)
		// 	fmt.Printf("Added %v by %v\n", release.Title, name)
		// }
	}
	// write the json
	// file, _ := os.OpenFile("big_encode.json", os.O_CREATE, os.ModePerm)
	// defer file.Close()
	// encoder := json.NewEncoder(file)
	// encoder.Encode(artists)
}
