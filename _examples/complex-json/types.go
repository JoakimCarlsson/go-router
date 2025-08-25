package main

type PanelRequest struct {
	PanelID         string `json:"panelId"`
	PanelItemOffset int    `json:"panelItemOffset"`
	PanelItemLimit  int    `json:"panelItemLimit"`
	SkipProgress    bool   `json:"skipProgress"`
}

type PanelResponse struct {
	Data PanelData `json:"data"`
}

type PanelData struct {
	Panel Panel `json:"panel"`
}

type Panel struct {
	Typename                 string           `json:"__typename"`
	ID                       string           `json:"id"`
	Title                    string           `json:"title"`
	DisplayHint              *DisplayHint     `json:"displayHint,omitempty"`
	Content                  *PanelContent    `json:"content,omitempty"`
	HideLabels               bool             `json:"hideLabels,omitempty"`
	Images                   *PanelImages     `json:"images,omitempty"`
	Pitch                    string           `json:"pitch,omitempty"`
	ShortPitch               string           `json:"shortPitch,omitempty"`
	LinkText                 string           `json:"linkText,omitempty"`
	Link                     interface{}      `json:"link,omitempty"`
	Trailers                 *Trailers        `json:"trailers,omitempty"`
	AutoPlayMedia            bool             `json:"autoPlayMedia,omitempty"`
	SinglePanels             []Panel          `json:"singlePanels,omitempty"`
	Subtitle                 string           `json:"subtitle,omitempty"`
	ShowMetadataForLink      bool             `json:"showMetadataForLink,omitempty"`
	HexColor                 string           `json:"hexColor,omitempty"`
	TierName                 string           `json:"tierName,omitempty"`
	Disclaimer               string           `json:"disclaimer,omitempty"`
	Bullets                  []string         `json:"bullets,omitempty"`
	CampaignDetails          *CampaignDetails `json:"campaignDetails,omitempty"`
	IsCompact                bool             `json:"isCompact,omitempty"`
	ShouldPlayVideo          bool             `json:"shouldPlayVideo,omitempty"`
	ShouldPlayVideoOutOfView bool             `json:"shouldPlayVideoOutOfView,omitempty"`
}

type DisplayHint struct {
	MediaPanelImageRatio string `json:"mediaPanelImageRatio"`
}

type PanelContent struct {
	PageInfo PageInfo    `json:"pageInfo"`
	Cards    []PanelCard `json:"cards"`
}

type PageInfo struct {
	HasNextPage    bool `json:"hasNextPage"`
	NextPageOffset int  `json:"nextPageOffset"`
	TotalCount     int  `json:"totalCount"`
}

type PanelCard struct {
	Typename                string                 `json:"__typename"`
	Series                  *Series                `json:"series,omitempty"`
	Movie                   *Movie                 `json:"movie,omitempty"`
	Clip                    *Clip                  `json:"clip,omitempty"`
	Short                   *Short                 `json:"short,omitempty"`
	Episode                 *Episode               `json:"episode,omitempty"`
	Channel                 *Channel               `json:"channel,omitempty"`
	SportEvent              *SportEvent            `json:"sportEvent,omitempty"`
	Page                    *PageReference         `json:"page,omitempty"`
	LinkText                string                 `json:"linkText,omitempty"`
	ContinueWatchingEntryID string                 `json:"continueWatchingEntryId,omitempty"`
	LabelText               string                 `json:"labelText,omitempty"`
	Progress                *Progress              `json:"progress,omitempty"`
	Media                   *ContinueWatchingMedia `json:"media,omitempty"`
}

type ContinueWatchingMedia struct {
	Typename   string      `json:"__typename"`
	Episode    *Episode    `json:"episode,omitempty"`
	Movie      *Movie      `json:"movie,omitempty"`
	SportEvent *SportEvent `json:"sportEvent,omitempty"`
}

type Progress struct {
	Percent  int    `json:"percent"`
	TimeLeft string `json:"timeLeft,omitempty"`
}

type CampaignDetails struct {
	CampaignLabelImage *Image `json:"campaignLabelImage"`
}

type PanelImages struct {
	Image16x9 *Image `json:"image16x9"`
	Image2x3  *Image `json:"image2x3"`
}

type Image struct {
	ID            string     `json:"id"`
	Source        string     `json:"source"`
	IsFallback    bool       `json:"isFallback,omitempty"`
	SourceEncoded string     `json:"sourceEncoded,omitempty"`
	Meta          *ImageMeta `json:"meta,omitempty"`
}

type ImageMeta struct {
	MuteBgColor *Color `json:"muteBgColor,omitempty"`
	BlurHash    string `json:"blurHash,omitempty"`
}

type Color struct {
	Hex string `json:"hex"`
}

type DateTime struct {
	ReadableDateTime string `json:"readableDateTime"`
	ReadableDate     string `json:"readableDate"`
	ReadableDateLong string `json:"readableDateLong"`
	ReadableDistance string `json:"readableDistance"`
	ISOString        string `json:"isoString"`
}

type Duration struct {
	ReadableShort string `json:"readableShort"`
	Seconds       int    `json:"seconds"`
}

type Synopsis struct {
	Brief    string `json:"brief"`
	Medium   string `json:"medium"`
	Long     string `json:"long"`
	Short    string `json:"short,omitempty"`
	Extended string `json:"extended,omitempty"`
}

type Label struct {
	ID                 string  `json:"id"`
	Airtime            *string `json:"airtime"`
	Announcement       *string `json:"announcement"`
	RecurringBroadcast string  `json:"recurringBroadcast"`
}

type ParentalRating struct {
	Sweden  *SwedenRating  `json:"sweden"`
	Finland *FinlandRating `json:"finland"`
}

type SwedenRating struct {
	AgeRecommendation   int  `json:"ageRecommendation"`
	SuitableForChildren bool `json:"suitableForChildren"`
}

type FinlandRating struct {
	AgeRestriction string `json:"ageRestriction"`
	Reason         string `json:"reason"`
}

type Upsell struct {
	PackageTierLink *PackageTierLink `json:"packageTierLink"`
	TierName        string           `json:"tierName"`
	LabelText       string           `json:"labelText"`
}

type PackageTierLink struct {
	TierID string `json:"tierId"`
}

type Access struct {
	HasAccess bool `json:"hasAccess"`
}

type Country struct {
	CountryCode string `json:"countryCode"`
	Name        string `json:"name"`
}

type Credits struct {
	Actors    []Actor  `json:"actors"`
	Directors []Credit `json:"directors"`
	Hosts     []Credit `json:"hosts"`
}

type Actor struct {
	CharacterName *string `json:"characterName"`
	Name          string  `json:"name"`
	Type          string  `json:"type"`
}

type Credit struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Trailers struct {
	MP4 string `json:"mp4"`
}

type Series struct {
	Typename                 string           `json:"__typename"`
	ID                       string           `json:"id"`
	Slug                     string           `json:"slug"`
	Title                    string           `json:"title"`
	Label                    *Label           `json:"label"`
	EditorialInfoText        *string          `json:"editorialInfoText"`
	Genres                   []string         `json:"genres"`
	MediaClassification      string           `json:"mediaClassification"`
	Synopsis                 *Synopsis        `json:"synopsis"`
	NumberOfAvailableSeasons int              `json:"numberOfAvailableSeasons"`
	Trailers                 *Trailers        `json:"trailers"`
	Images                   *SeriesImages    `json:"images"`
	Credits                  *Credits         `json:"credits"`
	ParentalRating           *ParentalRating  `json:"parentalRating"`
	IsPollFeatureEnabled     bool             `json:"isPollFeatureEnabled"`
	CDPPageOverride          *CDPPageOverride `json:"cdpPageOverride"`
	Upsell                   *Upsell          `json:"upsell"`
	AllSeasonLinks           []SeasonLink     `json:"allSeasonLinks,omitempty"`
}

type CDPPageOverride struct {
	ID string `json:"id"`
}

type SeriesImages struct {
	Main16x9          *Image `json:"main16x9"`
	Main16x9Annotated *Image `json:"main16x9Annotated"`
	Cover2x3          *Image `json:"cover2x3"`
	Poster2x3         *Image `json:"poster2x3"`
	BrandLogo         *Image `json:"brandLogo"`
	Logo              *Image `json:"logo"`
	RoundLogo         *Image `json:"roundLogo,omitempty"`
}

type SeasonLink struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	SeasonID         string `json:"seasonId"`
	NumberOfEpisodes int    `json:"numberOfEpisodes"`
}

type Movie struct {
	Typename               string          `json:"__typename"`
	ID                     string          `json:"id"`
	Slug                   string          `json:"slug"`
	Title                  string          `json:"title"`
	HumanCallToAction      string          `json:"humanCallToAction"`
	Label                  *Label          `json:"label"`
	EditorialInfoText      *string         `json:"editorialInfoText"`
	Genres                 []string        `json:"genres"`
	Synopsis               *Synopsis       `json:"synopsis"`
	Images                 *MovieImages    `json:"images"`
	PlayableFrom           *DateTime       `json:"playableFrom"`
	PlayableUntil          *DateTime       `json:"playableUntil"`
	IsLiveContent          bool            `json:"isLiveContent"`
	Access                 *Access         `json:"access"`
	Duration               *Duration       `json:"duration"`
	Trailers               *Trailers       `json:"trailers"`
	ProductionCountries    []Country       `json:"productionCountries"`
	ProductionYear         string          `json:"productionYear"`
	Credits                *MovieCredits   `json:"credits"`
	ParentalRating         *ParentalRating `json:"parentalRating"`
	Progress               *Progress       `json:"progress"`
	IsPollFeatureEnabled   bool            `json:"isPollFeatureEnabled"`
	MediaClassification    string          `json:"mediaClassification"`
	Upsell                 *Upsell         `json:"upsell"`
	OfflineDownloadAllowed bool            `json:"offlineDownloadAllowed"`
}

type MovieImages struct {
	Main16x9          *Image `json:"main16x9"`
	Main16x9Annotated *Image `json:"main16x9Annotated"`
	Cover2x3          *Image `json:"cover2x3"`
	Poster2x3         *Image `json:"poster2x3"`
	BrandLogo         *Image `json:"brandLogo"`
	Logo              *Image `json:"logo"`
}

type MovieCredits struct {
	Actors    []Actor  `json:"actors"`
	Directors []Credit `json:"directors"`
}

type Episode struct {
	Typename               string          `json:"__typename"`
	ID                     string          `json:"id"`
	Slug                   string          `json:"slug"`
	Title                  string          `json:"title"`
	ExtendedTitle          string          `json:"extendedTitle"`
	Synopsis               *Synopsis       `json:"synopsis"`
	SeasonID               string          `json:"seasonId"`
	Series                 *Series         `json:"series"`
	EpisodeNumber          int             `json:"episodeNumber"`
	PlayableFrom           *DateTime       `json:"playableFrom"`
	PlayableUntil          *DateTime       `json:"playableUntil"`
	LiveEventEnd           *DateTime       `json:"liveEventEnd"`
	IsLiveContent          bool            `json:"isLiveContent"`
	ParentalRating         *ParentalRating `json:"parentalRating"`
	Images                 *EpisodeImages  `json:"images"`
	IsPollFeatureEnabled   bool            `json:"isPollFeatureEnabled"`
	Progress               *Progress       `json:"progress"`
	Access                 *Access         `json:"access"`
	Duration               *Duration       `json:"duration"`
	MediaClassification    string          `json:"mediaClassification"`
	Upsell                 *Upsell         `json:"upsell"`
	OfflineDownloadAllowed bool            `json:"offlineDownloadAllowed"`
	IsStartOverEnabled     bool            `json:"isStartOverEnabled"`
}

type EpisodeImages struct {
	Main16x9          *Image `json:"main16x9"`
	Main16x9Annotated *Image `json:"main16x9Annotated"`
}

type Clip struct {
	Typename             string      `json:"__typename"`
	ID                   string      `json:"id"`
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	Slug                 string      `json:"slug"`
	Images               *ClipImages `json:"images"`
	IsPollFeatureEnabled bool        `json:"isPollFeatureEnabled"`
	PlayableFrom         *DateTime   `json:"playableFrom"`
	PlayableUntil        *DateTime   `json:"playableUntil"`
	Parent               *ClipParent `json:"parent"`
	MediaClassification  string      `json:"mediaClassification"`
	Access               *Access     `json:"access"`
	Duration             *Duration   `json:"duration"`
	IsLiveContent        bool        `json:"isLiveContent"`
}

type ClipImages struct {
	Main16x9 *Image `json:"main16x9"`
}

type ClipParent struct {
	Typename string `json:"__typename"`
	ID       string `json:"id"`
	Title    string `json:"title"`
}

type Short struct {
	Typename     string       `json:"__typename"`
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Slug         string       `json:"slug"`
	PlayableFrom *DateTime    `json:"playableFrom"`
	Duration     *Duration    `json:"duration"`
	Images       *ShortImages `json:"images"`
	LiveEpisode  *LiveEpisode `json:"liveEpisode"`
}

type ShortImages struct {
	FirstFrame9x16 *Image `json:"firstFrame9x16"`
	FirstFrame16x9 *Image `json:"firstFrame16x9"`
}

type LiveEpisode struct {
	ID     string             `json:"id"`
	Images *LiveEpisodeImages `json:"images"`
	Series *LiveEpisodeSeries `json:"series"`
}

type LiveEpisodeImages struct {
	RoundLogo *Image `json:"roundLogo"`
}

type LiveEpisodeSeries struct {
	Title string `json:"title"`
}

type SportEvent struct {
	Typename             string       `json:"__typename"`
	ID                   string       `json:"id"`
	Slug                 string       `json:"slug"`
	Title                string       `json:"title"`
	Genre                string       `json:"genre"`
	League               string       `json:"league"`
	Season               string       `json:"season"`
	Round                string       `json:"round"`
	Commentators         string       `json:"commentators"`
	Country              string       `json:"country"`
	Arena                string       `json:"arena"`
	Studio               string       `json:"studio"`
	InStudio             string       `json:"inStudio"`
	ProductionYear       string       `json:"productionYear"`
	IsLiveContent        bool         `json:"isLiveContent"`
	PlayableFrom         *DateTime    `json:"playableFrom"`
	PlayableUntil        *DateTime    `json:"playableUntil"`
	LiveEventEnd         *DateTime    `json:"liveEventEnd"`
	Synopsis             *Synopsis    `json:"synopsis"`
	Images               *SportImages `json:"images"`
	Trailers             *Trailers    `json:"trailers"`
	Access               *Access      `json:"access"`
	Upsell               *Upsell      `json:"upsell"`
	IsStartOverEnabled   bool         `json:"isStartOverEnabled"`
	EditorialInfoText    *string      `json:"editorialInfoText"`
	HumanCallToAction    string       `json:"humanCallToAction"`
	Progress             *Progress    `json:"progress"`
	IsPollFeatureEnabled bool         `json:"isPollFeatureEnabled"`
}

type SportImages struct {
	Main16x9          *Image `json:"main16x9"`
	Main16x9Annotated *Image `json:"main16x9Annotated"`
	Cover2x3          *Image `json:"cover2x3"`
	Poster2x3         *Image `json:"poster2x3"`
	Logo              *Image `json:"logo"`
	BrandLogo         *Image `json:"brandLogo"`
}

type Channel struct {
	Typename             string         `json:"__typename"`
	ID                   string         `json:"id"`
	Title                string         `json:"title"`
	Tagline              string         `json:"tagline"`
	Description          string         `json:"description"`
	Access               *Access        `json:"access"`
	Upsell               *Upsell        `json:"upsell"`
	Type                 string         `json:"type"`
	Images               *ChannelImages `json:"images"`
	EPG                  *EPGItem       `json:"epg"`
	IsDRMProtected       bool           `json:"isDrmProtected"`
	IsPollFeatureEnabled bool           `json:"isPollFeatureEnabled"`
}

type ChannelImages struct {
	Main16x9 *Image `json:"main16x9"`
	Logo     *Image `json:"logo"`
}

type EPGItem struct {
	Title    string       `json:"title"`
	Synopsis *EPGSynopsis `json:"synopsis"`
	Start    string       `json:"start"`
	End      string       `json:"end"`
	Type     string       `json:"type"`
	Images   *EPGImages   `json:"images"`
	Media    *EPGMedia    `json:"media"`
}

type EPGSynopsis struct {
	Brief    string `json:"brief"`
	Medium   string `json:"medium"`
	Long     string `json:"long"`
	Short    string `json:"short"`
	Extended string `json:"extended"`
}

type EPGImages struct {
	Main16x9 *Image `json:"main16x9"`
}

type EPGMedia struct {
	Typename   string      `json:"__typename"`
	Episode    *Episode    `json:"episode,omitempty"`
	Movie      *Movie      `json:"movie,omitempty"`
	SportEvent *SportEvent `json:"sportEvent,omitempty"`
}

type PageReference struct {
	Typename string      `json:"__typename"`
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Type     string      `json:"type,omitempty"`
	Images   *PageImages `json:"images"`
}

type PageImages struct {
	Image16x9 *Image `json:"image16x9"`
	Logo      *Image `json:"logo"`
}
