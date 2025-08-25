package main

type CarShowRequest struct {
	ShowID     string `json:"showId"`
	Offset     int    `json:"offset"`
	Limit      int    `json:"limit"`
	FilterType string `json:"filterType"`
}

type CarShowResponse struct {
	Data      CarShowData      `json:"data"`
	Metadata  ResponseMetadata `json:"metadata"`
	Relations *CarRelations    `json:"relations,omitempty"`
}

type CarShowData struct {
	CarShow       CarShow       `json:"carShow"`
	FeaturedCars  []Car         `json:"featuredCars"`
	Manufacturers []Manufacturer `json:"manufacturers"`
}

type ResponseMetadata struct {
	RequestID    string       `json:"requestId"`
	Timestamp    string       `json:"timestamp"`
	RelatedShows []CarShow    `json:"relatedShows,omitempty"`
	UserProfile  *UserProfile `json:"userProfile,omitempty"`
}

type CarRelations struct {
	CompetingCars []Car        `json:"competingCars"`
	SimilarShows  []CarShow    `json:"similarShows"`
	UserFavorites *UserProfile `json:"userFavorites,omitempty"`
}

type CarShow struct {
	ID               string           `json:"id"`
	Name             string           `json:"name"`
	Location         string           `json:"location"`
	Date             string           `json:"date"`
	Description      string           `json:"description"`
	FeaturedCars     []Car            `json:"featuredCars"`
	Sponsors         []Sponsor        `json:"sponsors"`
	RelatedShows     []CarShow        `json:"relatedShows,omitempty"`
	Images           *ShowImages      `json:"images,omitempty"`
	Organizer        *Organizer       `json:"organizer,omitempty"`
	CompetitionRules *CompetitionInfo `json:"competitionRules,omitempty"`
	IsActive         bool             `json:"isActive"`
	MaxCapacity      int              `json:"maxCapacity"`
	TicketPrice      float64          `json:"ticketPrice"`
	Theme            string           `json:"theme"`
	ChildShows       []CarShow        `json:"childShows,omitempty"`
}

type Car struct {
	ID           string        `json:"id"`
	Make         string        `json:"make"`
	Model        string        `json:"model"`
	Year         int           `json:"year"`
	Color        string        `json:"color"`
	Engine       *Engine       `json:"engine,omitempty"`
	Owner        *CarOwner     `json:"owner,omitempty"`
	Manufacturer *Manufacturer `json:"manufacturer,omitempty"`
	Images       *CarImages    `json:"images,omitempty"`
	Specs        *CarSpecs     `json:"specs,omitempty"`
	History      *CarHistory   `json:"history,omitempty"`
	RelatedCars  []Car         `json:"relatedCars,omitempty"`
	Shows        []CarShow     `json:"shows,omitempty"`
	IsVintage    bool          `json:"isVintage"`
	IsElectric   bool          `json:"isElectric"`
	Price        float64       `json:"price"`
}

type Manufacturer struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Country       string            `json:"country"`
	Founded       int               `json:"founded"`
	Models        []Car             `json:"models"`
	ParentCompany *Manufacturer     `json:"parentCompany,omitempty"`
	Subsidiaries  []Manufacturer    `json:"subsidiaries,omitempty"`
	History       *CompanyHistory   `json:"history,omitempty"`
	Images        *CompanyImages    `json:"images,omitempty"`
	IsActive      bool              `json:"isActive"`
	Partnerships  []Partnership     `json:"partnerships,omitempty"`
}

type UserProfile struct {
	ID           string      `json:"id"`
	Username     string      `json:"username"`
	Email        string      `json:"email"`
	FavoriteCars []Car       `json:"favoriteCars"`
	FavoriteShows []CarShow  `json:"favoriteShows"`
	OwnedCars    []Car       `json:"ownedCars,omitempty"`
	Friends      []UserProfile `json:"friends,omitempty"`
	Preferences  *UserPreferences `json:"preferences,omitempty"`
	CreatedAt    string      `json:"createdAt"`
	LastActive   string      `json:"lastActive"`
}

type Engine struct {
	Type         string  `json:"type"`
	Displacement float64 `json:"displacement"`
	Horsepower   int     `json:"horsepower"`
	Torque       int     `json:"torque"`
	FuelType     string  `json:"fuelType"`
	Cylinders    int     `json:"cylinders"`
	UsedInCars   []Car   `json:"usedInCars,omitempty"`
}

type CarOwner struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Email      string      `json:"email"`
	Location   string      `json:"location"`
	OwnedCars  []Car       `json:"ownedCars"`
	Profile    *UserProfile `json:"profile,omitempty"`
	JoinedDate string      `json:"joinedDate"`
}

type Sponsor struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Website      string    `json:"website"`
	Logo         *Image    `json:"logo,omitempty"`
	SponsoredShows []CarShow `json:"sponsoredShows,omitempty"`
	SponsorLevel string    `json:"sponsorLevel"`
	Budget       float64   `json:"budget"`
}

type Organizer struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	OrganizedShows []CarShow `json:"organizedShows,omitempty"`
	Experience    int       `json:"experience"`
	Rating        float64   `json:"rating"`
}

type Partnership struct {
	ID           string        `json:"id"`
	PartnerA     *Manufacturer `json:"partnerA"`
	PartnerB     *Manufacturer `json:"partnerB"`
	Type         string        `json:"type"`
	StartDate    string        `json:"startDate"`
	EndDate      string        `json:"endDate,omitempty"`
	Description  string        `json:"description"`
	IsActive     bool          `json:"isActive"`
}

type CompetitionInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Rules       []string  `json:"rules"`
	Categories  []string  `json:"categories"`
	Prizes      []Prize   `json:"prizes"`
	Judges      []Judge   `json:"judges"`
	StartTime   string    `json:"startTime"`
	EndTime     string    `json:"endTime"`
	Show        *CarShow  `json:"show,omitempty"`
}

type Prize struct {
	Position    int     `json:"position"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Sponsor     *Sponsor `json:"sponsor,omitempty"`
}

type Judge struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Experience int       `json:"experience"`
	Specialty  string    `json:"specialty"`
	Events     []CarShow `json:"events,omitempty"`
}

type UserPreferences struct {
	FavoriteMakes      []string `json:"favoriteMakes"`
	PreferredYearRange *YearRange `json:"preferredYearRange,omitempty"`
	MaxPrice          float64  `json:"maxPrice"`
	NotificationSettings *NotificationSettings `json:"notificationSettings,omitempty"`
	User              *UserProfile `json:"user,omitempty"`
}

type YearRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type NotificationSettings struct {
	EmailEnabled  bool `json:"emailEnabled"`
	SMSEnabled    bool `json:"smsEnabled"`
	PushEnabled   bool `json:"pushEnabled"`
	Frequency     string `json:"frequency"`
}

type CarImages struct {
	Front    *Image `json:"front,omitempty"`
	Rear     *Image `json:"rear,omitempty"`
	Side     *Image `json:"side,omitempty"`
	Interior *Image `json:"interior,omitempty"`
	Engine   *Image `json:"engine,omitempty"`
	Gallery  []Image `json:"gallery,omitempty"`
}

type ShowImages struct {
	Banner    *Image  `json:"banner,omitempty"`
	Thumbnail *Image  `json:"thumbnail,omitempty"`
	Gallery   []Image `json:"gallery,omitempty"`
}

type CompanyImages struct {
	Logo      *Image  `json:"logo,omitempty"`
	Headquarters *Image `json:"headquarters,omitempty"`
	Gallery   []Image `json:"gallery,omitempty"`
}

type Image struct {
	ID          string `json:"id"`
	URL         string `json:"url"`
	Alt         string `json:"alt"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	Caption     string `json:"caption,omitempty"`
	Photographer string `json:"photographer,omitempty"`
}

type CarSpecs struct {
	Length       float64 `json:"length"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Weight       int     `json:"weight"`
	TopSpeed     int     `json:"topSpeed"`
	Acceleration string  `json:"acceleration"`
	FuelEconomy  string  `json:"fuelEconomy"`
	Drivetrain   string  `json:"drivetrain"`
	Transmission string  `json:"transmission"`
}

type CarHistory struct {
	PreviousOwners []CarOwner `json:"previousOwners"`
	Accidents      []Accident `json:"accidents,omitempty"`
	Modifications  []Modification `json:"modifications,omitempty"`
	ServiceRecords []ServiceRecord `json:"serviceRecords,omitempty"`
	Mileage        int        `json:"mileage"`
	Car            *Car       `json:"car,omitempty"`
}

type Accident struct {
	Date        string `json:"date"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Cost        float64 `json:"cost"`
	Images      []Image `json:"images,omitempty"`
}

type Modification struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	InstallDate string  `json:"installDate"`
	Installer   string  `json:"installer"`
}

type ServiceRecord struct {
	Date        string  `json:"date"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Cost        float64 `json:"cost"`
	ServiceShop string  `json:"serviceShop"`
	Mileage     int     `json:"mileage"`
}

type CompanyHistory struct {
	Founded       int       `json:"founded"`
	Founder       string    `json:"founder"`
	Milestones    []Milestone `json:"milestones"`
	Acquisitions  []Acquisition `json:"acquisitions,omitempty"`
	Company       *Manufacturer `json:"company,omitempty"`
}

type Milestone struct {
	Year        int    `json:"year"`
	Event       string `json:"event"`
	Description string `json:"description"`
}

type Acquisition struct {
	Year        int           `json:"year"`
	Company     string        `json:"company"`
	Amount      float64       `json:"amount"`
	Reason      string        `json:"reason"`
	Acquirer    *Manufacturer `json:"acquirer,omitempty"`
}