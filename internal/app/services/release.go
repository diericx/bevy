package services

import (
	"fmt"
	"log"
	"sort"

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

type ScoredRelease struct {
	app.Release
	SeederScore  float64
	SizeScore    float64
	QualityScore float64
}

func (s *Release) QueryMovie(imdbID string, title string, year string) ([]ScoredRelease, error) {
	scoredReleases := make([]ScoredRelease, 0)

	sizeScoreExpression, err := govaluate.NewEvaluableExpression(s.Config.SizeScoreFunc)
	if err != nil {
		return nil, fmt.Errorf("Error evaluating size score function: %v", err)
	}
	seederScoreExpression, err := govaluate.NewEvaluableExpression(s.Config.SeederScoreFunc)
	if err != nil {
		return nil, fmt.Errorf("Error evaluating seeder score function: %v", err)
	}
	qualityScoreExpression, err := govaluate.NewEvaluableExpression(s.Config.QualityScoreFunc)
	if err != nil {
		return nil, fmt.Errorf("Error evaluating seeder score function: %v", err)
	}

	for qualityIndex, quality := range s.Config.Qualities {
		releases, err := s.ReleaseRepo.QueryAllIndexers(
			imdbID,
			fmt.Sprintf("%s %s %s", title, year, quality.Regex),
		)
		if err != nil {
			return nil, err
		}

		// Convert releases to scored release
		for _, release := range releases {
			if release.ValidateSizeSeedersAndName(quality) != nil {
				continue
			}
			params := map[string]interface{}{
				"seeders": release.Seeders,
				"quality": qualityIndex,
				"sizeMB":  release.Size / 1024 / 1024,
				"sizeGB":  release.Size / 1024 / 1024 / 1024,
			}
			seederScore, err := seederScoreExpression.Evaluate(params)
			if err != nil {
				log.Fatalf("Error evaluating seeder sorting function %s \n %v", s.Config.SeederScoreFunc, err)
				return nil, err
			}
			sizeScore, err := sizeScoreExpression.Evaluate(params)
			if err != nil {
				log.Fatalf("Error evaluating size sorting function %s \n %v", s.Config.SizeScoreFunc, err)
				return nil, err
			}
			qualityScore, err := qualityScoreExpression.Evaluate(params)
			if err != nil {
				log.Fatalf("Error evaluating size sorting function %s \n %v", s.Config.SizeScoreFunc, err)
				return nil, err
			}

			scoredReleases = append(scoredReleases, ScoredRelease{
				Release:      release,
				SeederScore:  seederScore.(float64),
				SizeScore:    sizeScore.(float64),
				QualityScore: qualityScore.(float64),
			})
		}
	}

	sortScoredReleases(scoredReleases)

	return scoredReleases, nil
}

// sortScoredReleases will sort the given slice of scored releases by their cum score
func sortScoredReleases(scoredReleases []ScoredRelease) {
	sort.Slice(scoredReleases, func(i, j int) bool {
		iScore := scoredReleases[i].SeederScore + scoredReleases[i].SizeScore + scoredReleases[i].QualityScore
		jScore := scoredReleases[j].SeederScore + scoredReleases[j].SizeScore + scoredReleases[j].QualityScore

		return iScore > jScore
	})
}
