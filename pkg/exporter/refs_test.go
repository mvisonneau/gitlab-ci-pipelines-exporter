package exporter

// TODO: Reimplement these tests
// Here an example of concurrent execution of projects polling
// func TestProjectPolling(t *testing.T) {
// 	projects := []schemas.Project{{Name: "test1"}, {Name: "test2"}, {Name: "test3"}, {Name: "test4"}}
// 	until := make(chan struct{})
// 	defer close(until)
// 	_, _, c := getMockedClient()
// 	// provided we are able to intercept an error from from pollProject method
// 	// we can iterate over a channel of Project and collect its results
// 	assert.Equal(t, len(projects), pollingResult(until, readProjects(until, projects...), c, t))
// }

// func pollingResult(until <-chan struct{}, projects <-chan schemas.Project, client *Client, t *testing.T) (numErrs int) {
// 	for i := range projects {
// 		select {
// 		case <-until:
// 			return numErrs
// 		default:
// 			if assert.Error(t, client.pollProject(i)) {
// 				numErrs++
// 			}
// 		}
// 	}
// 	return numErrs
// }

// func TestPollProjectsRefs(t *testing.T) {
// 	message := "some error"
// 	doing := func() func(*ProjectRef) error {
// 		return func(*ProjectRef) error {
// 			// set the already polled refs, simulate the pollProject(p Project) set of Client.hasPolledOnInit
// 			// return an error to count them afterwards
// 			return fmt.Errorf(message)
// 		}
// 	}
// 	testProjects := ProjectsRefs{}
// 	testProjects[1] = map[string]*ProjectRef{"master": &ProjectRef{}}
// 	testProjects[2] = map[string]*ProjectRef{"master": &ProjectRef{}}

// 	until := make(chan struct{})
// 	errCh := pollProjectsRefs(2, doing(), until, testProjects)
// 	var errCount int
// 	for err := range errCh {
// 		if assert.Error(t, err) {
// 			assert.Equal(t, err.Error(), message)
// 			errCount++
// 		}
// 	}
// 	assert.Equal(t, len(testProjects), errCount)
// }
