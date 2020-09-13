package app

// func AddBestRelease(torrentService TorrentService, releases []Release, q Quality, minSeeders int) error {
// 	// sort releases by seeders (to get most available torrents first)
// 	sort.Slice(releases, func(i, j int) bool {
// 		return releases[i].Seeders > releases[j].Seeders
// 	})

// 	for _, t := range releases {
// 		if t.Size < q.MinSize || t.Size > q.MaxSize {
// 			log.Printf("INFO: Passing on release %s because size %v is not correct.", t.Title, t.Size)
// 			continue
// 		}
// 		if t.Seeders < minSeeders {
// 			log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", t.Title, t.Seeders, minSeeders)
// 			continue
// 		}
// 		if StringContainsAnyOf(strings.ToLower(t.Title), GetBlacklistedTorrentNameContents()) {
// 			log.Printf("INFO: Passing on release %s because title contains one of these blacklisted words: %+v", t.Title, GetBlacklistedTorrentNameContents())
// 			continue
// 		}

// 		// Add to client to get hash
// 		torrentAdded, err := torrentService.AddFromURLUknownScheme(t.Link, t.LinkAuth)
// 		if err != nil {
// 			log.Printf("WARNING: could not add torrent magnet for %s\n Err: %s", t.Title, err)
// 			continue
// 		}

// 		// Attempt to find a valid file
// 		index, terr := findMovieFileInTorrent(torrentService, torrentAdded)
// 		if terr != nil {
// 			log.Printf("INFO: Passing on release %s because there was no valid file.", t.Title)
// 			continue
// 		}
// 		// TODO: Do better logic here to find "main" file
// 		t.MainFileIndex = index

// 		copyT := t // Why does this object need to be copied?? So weird..
// 		return &copyT, nil
// 	}
// 	return nil, nil
// }

// func findMovieFileInTorrent(s TorrentService, t *Torrent) (int, error) {
// 	// Get correct file
// 	files, err := s.GetFiles(t)
// 	if err != nil {
// 		return 0, err
// 	}

// 	// Sort by size
// 	sort.Slice(files, func(i, j int) bool {
// 		return files[i].Size > files[j].Size
// 	})

// 	for i, file := range files {
// 		if StringEndsInAny(strings.ToLower(file.Path), GetSupportedVideoFileFormats()) && !StringContainsAnyOf(strings.ToLower(file.Path), GetBlacklistedFileNameContents()) {
// 			return i, nil
// 		}
// 	}
// 	return 0, errors.New("no valid file in torrent")
// }

// func StringContainsAnyOf(s string, substrings []string) bool {
// 	for _, substring := range substrings {
// 		if strings.Contains(s, substring) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func StringEndsInAny(s string, suffixes []string) bool {
// 	for _, suffix := range suffixes {
// 		if strings.HasSuffix(s, suffix) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func ParseTorrentIdFromString(id string) (int, *Error) {
// 	idInt64, err := strconv.ParseInt(id, 10, 32)
// 	if err != nil {
// 		return 0, NewError(err, http.StatusBadRequest, InvalidIDErr)
// 	}
// 	return int(idInt64), nil
// }
