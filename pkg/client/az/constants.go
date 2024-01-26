package az

const DefaultDetectionApi string = "language/:analyze-text?api-version=2022-05-01"
const DefaultLanguage string = "en"

const ShowStatsParam string = "&showStats=true"

func GetDefaultDetectionApi(showStats bool) string {
	if showStats {
		return DefaultDetectionApi + ShowStatsParam
	}
	return DefaultDetectionApi
}
