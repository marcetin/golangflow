package actions_test

func (as *ActionSuite) Test_HomeHandler() {
	res := as.HTML("/").Get()

	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), "Golangflow")
}
