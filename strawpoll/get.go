package strawpoll

import "time"

const strawpollGetEndpoint = "https://strawpoll.com/api/poll/"

type StrawPollGetClient interface {
	GetPoll(ID string) (*StrawPollResults, error)
}

type StrawPollResults struct {
	Content Content
}

type Content struct {
	Comments         int
	CreatedAt        time.Time `json:"created_at"`
	Creator          Creator
	Deadline         time.Time
	HasWebhooks      int 	   `json:"has_webhooks"`
	ID               string
	Media            Media
	OriginalDeadline time.Time `json:"original_deadline"`
	Pin              string
	Poll             Poll
	Status           string
	Title            string
	Type             string
}

type Creator struct {
	AvatarPath    string `json:"avatar_path"`
	DisplayName   string
	MonthlyPoints int 	 `json:"monthly_points"`
	Username      string
}

type Media struct {
	Hash   string
	Path   string
	Unused int
}

type Poll struct {
	IsPointsEligible int           `json:"is_points_eligible"`
	IsVotable        int          `json:"is_votable"`
	LastVoteAt       time.Time     `json:"last_vote_at"`
	OriginalTitle    string        `json:"original_title"`
	PollAnswers      []PollAnswer  `json:"poll_answers"`
	PollInfo         PollInfo      `json:"poll_info"`
	Private          int
	Title            string
	TotalVoters      int 		   `json:"total_voters"`
	TotalVotes       int 		   `json:"total_votes"`
}

type PollAnswer struct {
	Answer         string
	ID             string
	OriginalAnswer string `json:"original_answer"`
	Sorting        int
	Type           string
	Votes          int
}

type PollInfo struct {
	Captcha             int
	CreatorCountry      string 	  `json:"creator_country"`
	Description         string
	EditedAt            time.Time `json:"edited_at"`
	OriginalDescription string    `json:"original_description"`
	ShowResults         int       `json:"show_results"`
}