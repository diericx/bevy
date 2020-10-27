package services

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/diericx/iceetime/internal/app"
	"github.com/diericx/iceetime/internal/app/repos/jackett"
)

// Release service uses the jackett repo along with configurable sorting functions in order to provide a
// quality list of torrents that should be downloaded for any given query.
type Release struct {
	ReleaseRepo jackett.ReleaseRepo
	Config      app.ReleaseServiceConfig
}

func (s *Release) QueryMovie(imdbID string, title string, year string, minQuality int) ([]app.Release, error) {
	quality := s.Config.Qualities[minQuality]
	releases, err := s.ReleaseRepo.QueryAllIndexers(
		imdbID,
		fmt.Sprintf("%s %s %s", title, year, quality.Regex),
	)
	if err != nil {
		return nil, err
	}

	releases = s.getOnlyValidReleasesBasedOnSizeSeedersAndName(releases, quality)

	err = s.sortReleasesViaSizeAndSeederFunctions(releases)

	return releases, err
}

// sortReleasesViaSizeAndSeederFunctions will sort the given slice using functions from config
func (s *Release) sortReleasesViaSizeAndSeederFunctions(releases []app.Release) error {
	sizeScoreExpression, err := govaluate.NewEvaluableExpression(s.Config.SizeScoreFunc)
	if err != nil {
		return fmt.Errorf("Error evaluating size score function: %v", err)
	}
	seederScoreExpression, err := govaluate.NewEvaluableExpression(s.Config.SeederScoreFunc)
	if err != nil {
		return fmt.Errorf("Error evaluating seeder score function: %v", err)
	}

	// sort torrents by seeder and size functions
	sort.Slice(releases, func(i, j int) bool {
		log.Printf("Comparing: \n %+v \n %+v", releases[i], releases[j])
		iParams := map[string]interface{}{
			"seeders": releases[i].Seeders,
			"sizeGB":  releases[i].Size / 1024 / 1024 / 1024,
		}
		jParams := map[string]interface{}{
			"seeders": releases[j].Seeders,
			"sizeGB":  releases[j].Size / 1024 / 1024 / 1024,
		}

		iSeederScore, err := seederScoreExpression.Evaluate(iParams)
		if err != nil {
			log.Fatalf("Error evaluating seeder sorting function %s \n %v", s.Config.SeederScoreFunc, err)
			err = err
			return false
		}
		iSizeScore, err := sizeScoreExpression.Evaluate(iParams)
		if err != nil {
			log.Fatalf("Error evaluating size sorting function %s \n %v", s.Config.SizeScoreFunc, err)
			err = err
			return false
		}
		jSeederScore, err := seederScoreExpression.Evaluate(jParams)
		if err != nil {
			log.Fatalf("Error evaluating seeder sorting function %s \n %v", s.Config.SeederScoreFunc, err)
			err = err
			return false
		}
		jSizeScore, err := sizeScoreExpression.Evaluate(jParams)
		if err != nil {
			log.Fatalf("Error evaluating size sorting function %s \n %v", s.Config.SizeScoreFunc, err)
			err = err
			return false
		}

		iScore := iSeederScore.(float64) + iSizeScore.(float64)
		jScore := jSeederScore.(float64) + jSizeScore.(float64)

		log.Printf("Scores: \n seeders: %v, size: %v, \n seeders: %v, size: %v", iSeederScore.(float64), iSizeScore.(float64), jSeederScore.(float64), jSizeScore.(float64))

		return iScore > jScore
	})

	return err
}

// getOnlyValidReleasesBasedOnSizeSeedersAndName will trim the given release slice and return only those that are valid given
// size min maxes, seeder min maxes and name blacklists.
func (s *Release) getOnlyValidReleasesBasedOnSizeSeedersAndName(releases []app.Release, quality app.Quality) []app.Release {
	validReleases := []app.Release{}
	for _, release := range releases {
		if float64(release.Size) < quality.MinSize || float64(release.Size) > quality.MaxSize {
			log.Printf("INFO: Passing on release %s because size %v is not correct.", release.Title, release.Size)
			continue
		}
		if release.Seeders < int(quality.MinSeeders) {
			log.Printf("INFO: Passing on release %s because seeders: %v is less than minimum: %v", release.Title, release.Seeders, quality.MinSeeders)
			continue
		}
		if stringContainsAnyOf(strings.ToLower(release.Title), app.GetBlacklistedTorrentNameContents()) {
			log.Printf("INFO: Passing on release %s because title contains one of these blacklisted words: %+v", release.Title, app.GetBlacklistedTorrentNameContents())
			continue
		}
		validReleases = append(validReleases, release)
	}
	return validReleases
}
