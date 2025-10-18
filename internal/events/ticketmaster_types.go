package events

import "time"

type TicketmasterEventSearchResponse struct {
	Embedded struct {
		Events []TicketmasterEvent `json:"events"`
	} `json:"_embedded"`
}

type TicketmasterEvent struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Id     string `json:"id"`
	Test   bool   `json:"test"`
	Url    string `json:"url,omitempty"`
	Locale string `json:"locale"`
	Images []struct {
		Ratio    string `json:"ratio"`
		Url      string `json:"url"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Fallback bool   `json:"fallback"`
	} `json:"images"`
	Sales struct {
		Public struct {
			StartDateTime time.Time `json:"startDateTime"`
			StartTBD      bool      `json:"startTBD"`
			StartTBA      bool      `json:"startTBA"`
			EndDateTime   time.Time `json:"endDateTime"`
		} `json:"public"`
		Presales []struct {
			StartDateTime time.Time `json:"startDateTime"`
			EndDateTime   time.Time `json:"endDateTime"`
			Name          string    `json:"name"`
		} `json:"presales"`
	} `json:"sales,omitempty"`
	Dates struct {
		Start struct {
			LocalDate      string    `json:"localDate"`
			LocalTime      string    `json:"localTime"`
			DateTime       time.Time `json:"dateTime"`
			DateTBD        bool      `json:"dateTBD"`
			DateTBA        bool      `json:"dateTBA"`
			TimeTBA        bool      `json:"timeTBA"`
			NoSpecificTime bool      `json:"noSpecificTime"`
		} `json:"start"`
		Timezone string `json:"timezone"`
		Status   struct {
			Code string `json:"code"`
		} `json:"status"`
		SpanMultipleDays bool `json:"spanMultipleDays"`
		End              struct {
			LocalDate      string    `json:"localDate"`
			LocalTime      string    `json:"localTime"`
			DateTime       time.Time `json:"dateTime"`
			Approximate    bool      `json:"approximate"`
			NoSpecificTime bool      `json:"noSpecificTime"`
		} `json:"end,omitempty"`
	} `json:"dates"`
	Classifications []struct {
		Primary bool `json:"primary"`
		Segment struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"segment"`
		Genre struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"genre"`
		SubGenre struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"subGenre"`
		Type struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"type"`
		SubType struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"subType"`
		Family bool `json:"family"`
	} `json:"classifications,omitempty"`
	Promoter struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"promoter,omitempty"`
	Promoters []struct {
		Id          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"promoters,omitempty"`
	Info        string `json:"info,omitempty"`
	PleaseNote  string `json:"pleaseNote,omitempty"`
	PriceRanges []struct {
		Type     string  `json:"type"`
		Currency string  `json:"currency"`
		Min      float64 `json:"min"`
		Max      float64 `json:"max"`
	} `json:"priceRanges,omitempty"`
	Seatmap struct {
		StaticUrl string `json:"staticUrl"`
		Id        string `json:"id"`
	} `json:"seatmap,omitempty"`
	Accessibility struct {
		TicketLimit int    `json:"ticketLimit"`
		Id          string `json:"id"`
	} `json:"accessibility,omitempty"`
	TicketLimit struct {
		Info string `json:"info"`
		Id   string `json:"id"`
	} `json:"ticketLimit,omitempty"`
	AgeRestrictions struct {
		LegalAgeEnforced bool   `json:"legalAgeEnforced"`
		Id               string `json:"id"`
	} `json:"ageRestrictions,omitempty"`
	Ticketing struct {
		SafeTix struct {
			Enabled          bool `json:"enabled"`
			InAppOnlyEnabled bool `json:"inAppOnlyEnabled,omitempty"`
		} `json:"safeTix"`
		AllInclusivePricing struct {
			Enabled bool `json:"enabled"`
		} `json:"allInclusivePricing,omitempty"`
		Id string `json:"id"`
	} `json:"ticketing"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		Attractions []struct {
			Href string `json:"href"`
		} `json:"attractions,omitempty"`
		Venues []struct {
			Href string `json:"href"`
		} `json:"venues"`
	} `json:"_links"`
	Embedded struct {
		Venues []struct {
			Name       string `json:"name"`
			Type       string `json:"type"`
			Id         string `json:"id"`
			Test       bool   `json:"test"`
			Url        string `json:"url,omitempty"`
			Locale     string `json:"locale"`
			PostalCode string `json:"postalCode"`
			Timezone   string `json:"timezone"`
			City       struct {
				Name string `json:"name"`
			} `json:"city"`
			State struct {
				Name      string `json:"name"`
				StateCode string `json:"stateCode"`
			} `json:"state"`
			Country struct {
				Name        string `json:"name"`
				CountryCode string `json:"countryCode"`
			} `json:"country"`
			Address struct {
				Line1 string `json:"line1"`
			} `json:"address"`
			Location struct {
				Longitude string `json:"longitude"`
				Latitude  string `json:"latitude"`
			} `json:"location"`
			Markets []struct {
				Name string `json:"name"`
				Id   string `json:"id"`
			} `json:"markets,omitempty"`
			Dmas []struct {
				Id int `json:"id"`
			} `json:"dmas,omitempty"`
			BoxOfficeInfo struct {
				PhoneNumberDetail     string `json:"phoneNumberDetail"`
				OpenHoursDetail       string `json:"openHoursDetail"`
				AcceptedPaymentDetail string `json:"acceptedPaymentDetail"`
				WillCallDetail        string `json:"willCallDetail"`
			} `json:"boxOfficeInfo,omitempty"`
			ParkingDetail string `json:"parkingDetail,omitempty"`
			GeneralInfo   struct {
				GeneralRule string `json:"generalRule"`
				ChildRule   string `json:"childRule"`
			} `json:"generalInfo,omitempty"`
			UpcomingEvents struct {
				Archtics     int `json:"archtics"`
				Tmr          int `json:"tmr,omitempty"`
				Ticketmaster int `json:"ticketmaster"`
				Total        int `json:"_total"`
				Filtered     int `json:"_filtered"`
			} `json:"upcomingEvents"`
			Ada struct {
				AdaPhones     string `json:"adaPhones"`
				AdaCustomCopy string `json:"adaCustomCopy"`
				AdaHours      string `json:"adaHours"`
			} `json:"ada,omitempty"`
			Links struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
		} `json:"venues"`
		Attractions []struct {
			Name          string `json:"name"`
			Type          string `json:"type"`
			Id            string `json:"id"`
			Test          bool   `json:"test"`
			Url           string `json:"url"`
			Locale        string `json:"locale"`
			ExternalLinks struct {
				Twitter []struct {
					Url string `json:"url"`
				} `json:"twitter"`
				Facebook []struct {
					Url string `json:"url"`
				} `json:"facebook"`
				Wiki []struct {
					Url string `json:"url"`
				} `json:"wiki"`
				Instagram []struct {
					Url string `json:"url"`
				} `json:"instagram"`
				Homepage []struct {
					Url string `json:"url"`
				} `json:"homepage"`
			} `json:"externalLinks"`
			Aliases []string `json:"aliases,omitempty"`
			Images  []struct {
				Ratio       string `json:"ratio"`
				Url         string `json:"url"`
				Width       int    `json:"width"`
				Height      int    `json:"height"`
				Fallback    bool   `json:"fallback"`
				Attribution string `json:"attribution,omitempty"`
			} `json:"images"`
			Classifications []struct {
				Primary bool `json:"primary"`
				Segment struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"segment"`
				Genre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"genre"`
				SubGenre struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subGenre"`
				Type struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"type"`
				SubType struct {
					Id   string `json:"id"`
					Name string `json:"name"`
				} `json:"subType"`
				Family bool `json:"family"`
			} `json:"classifications"`
			UpcomingEvents struct {
				Tmr          int `json:"tmr"`
				Ticketmaster int `json:"ticketmaster"`
				Total        int `json:"_total"`
				Filtered     int `json:"_filtered"`
			} `json:"upcomingEvents"`
			Links struct {
				Self struct {
					Href string `json:"href"`
				} `json:"self"`
			} `json:"_links"`
		} `json:"attractions,omitempty"`
	} `json:"_embedded"`
}
