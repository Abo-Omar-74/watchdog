package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	onebusaway "github.com/OneBusAway/go-sdk"
	"github.com/OneBusAway/go-sdk/option"
	"github.com/jamespfennell/gtfs"
	"watchdog.onebusaway.org/internal/models"
)

func CheckAgenciesWithCoverage(cachePath string, logger *slog.Logger, server models.ObaServer) (int, error) {
	file, err := os.Open(cachePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return 0, err
	}

	fileBytes := make([]byte, fileInfo.Size())
	_, err = file.Read(fileBytes)
	if err != nil {
		return 0, err
	}

	staticData, err := gtfs.ParseStatic(fileBytes, gtfs.ParseStaticOptions{})
	if err != nil {
		return 0, err
	}

	if len(staticData.Agencies) == 0 {
		return 0, fmt.Errorf("no agencies found in GTFS bundle")
	}

	AgenciesInStaticGtfs.WithLabelValues(
		strconv.Itoa(server.ID),
	).Set(float64(len(staticData.Agencies)))

	return len(staticData.Agencies), nil
}

func GetAgenciesWithCoverage(server models.ObaServer) (int, error) {
	client := onebusaway.NewClient(
		option.WithAPIKey(server.ObaApiKey),
		option.WithBaseURL(server.ObaBaseURL),
	)

	ctx := context.Background()

	response, err := client.AgenciesWithCoverage.List(ctx)

	if err != nil {
		return 0, err
	}

	AgenciesInCoverageEndpoint.WithLabelValues(
		strconv.Itoa(server.ID),
	).Set(float64(len(response.Data.List)))

	return len(response.Data.List), nil
}

func CheckAgenciesWithCoverageMatch(cachePath string, logger *slog.Logger, server models.ObaServer) error {
	staticGtfsAgenciesCount, err := CheckAgenciesWithCoverage(cachePath, logger, server)
	if err != nil {
		return err
	}

	coverageAgenciesCount, err := GetAgenciesWithCoverage(server)

	matchValue := 0
	if coverageAgenciesCount == staticGtfsAgenciesCount {
		matchValue = 1
	}

	AgenciesMatch.WithLabelValues(strconv.Itoa(server.ID)).Set(float64(matchValue))

	return nil
}
