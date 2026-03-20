package engine

import "searx-cli/internal/types"

type Result = types.Result

type SearchEngine interface {
	Search(query string) ([]types.Result, error)
	Name() string
}
