package strawpoll

const strawpollGetEndpoint = "https://api.strawpoll.com/v2/polls/"

type StrawPollGetClient interface {
	GetPoll(ID string) (*StrawPollResults, error)
}

type StrawPollResults struct {
	Poll Poll
}

type Poll struct {
	Comments         int
	CreatedAt        int64 		    `json:"created_at"`
	PollConfig		 PollConfig	    `json:"poll_config"`
	PollOptions      []PollOptions  `json:"poll_options"`
	HasWebhooks      int 	        `json:"has_webhooks"`
	ID               string
	OriginalDeadline int64 			`json:"original_deadline"`
	Pin              string
	Status           string
	Title            string
	Type             string
}

type PollConfig struct {
	DeadlineAt	int64 `json:"deadline_at"`
}

type PollOptions struct {
	Value          string
	ID             string
	MaxVotes	   int `json:"max_votes"`
	Position	   int
	VoteCount      int `json:"vote_count"`
}
