package gitlab

import (
	"github.com/xanzy/go-gitlab"
)

func Client(token string) (*gitlab.Client, error) {
	// git, err := gitlab.NewClient(viper.GetString("gitlab_token"))
	git, err := gitlab.NewClient(token)
	if err != nil {
		return nil, err
	}
	return git, nil
}
