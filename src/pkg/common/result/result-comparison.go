package result

import (
	"strconv"
	"strings"

	oscalTypes "github.com/defenseunicorns/go-oscal/src/types/oscal-1-1-3"
	"github.com/mike-winberry/lulalib/src/pkg/message"
)

type StateChange string

const (
	NOT_SATISFIED_TO_SATISFIED StateChange = "NOT SATISFIED TO SATISFIED"
	SATISFIED_TO_NOT_SATISFIED StateChange = "SATISFIED TO NOT SATISFIED"
	NEW                        StateChange = "NEW"
	REMOVED                    StateChange = "REMOVED"
	UNCHANGED                  StateChange = "UNCHANGED"
)

type ResultComparison struct {
	StateChange      StateChange
	Satisfied        bool
	Finding          *oscalTypes.Finding
	ComparedFinding  *oscalTypes.Finding
	ObservationPairs []*ObservationPair
}

// PrintResultComparisonTable prints a table output of compared results
func (r ResultComparison) PrintResultComparisonTable(changedOnly bool) error {
	header := []string{"Observation", "Satisfied", "Change", "New Remarks", "Threshold Remarks"}
	rows := make([][]string, 0)
	columnSize := []int{20, 10, 15, 25, 30}

	for _, observationPair := range r.ObservationPairs {
		if changedOnly && observationPair.StateChange == UNCHANGED {
			continue
		}

		rows = append(rows, []string{
			observationPair.Name,
			convertSatisfied(observationPair.Satisfied, observationPair.StateChange),
			string(observationPair.StateChange),
			observationPair.Observation,
			observationPair.ComparedObservation,
		})
	}
	if len(rows) != 0 {
		err := message.Table(header, rows, columnSize)
		if err != nil {
			return err
		}
	}
	return nil
}

type ResultComparisonMap map[string]ResultComparison

// PrintObservationComparisonTable prints a table output of compared observations, per control
func (rm ResultComparisonMap) PrintObservationComparisonTable(changedOnly bool, skipRemoved bool, failedOnly bool) ([]string, error) {
	header := []string{"Control ID(s)", "Observation", "Satisfied", "Change", "New Remarks", "Threshold Remarks"}
	rows := make([][]string, 0)
	columnSize := []int{10, 20, 5, 15, 25, 25}

	// map[string]ObservationPairs (observationUUIDs -> observationPairs), map[string][]string (observationUUIDs -> controlIds)
	observationPairMap, controlObservationMap, noObservations := RefactorObservationsByControls(rm)

	for id, observationPair := range observationPairMap {
		if changedOnly && observationPair.StateChange == UNCHANGED {
			continue
		}
		if skipRemoved && observationPair.StateChange == REMOVED {
			continue
		}
		if failedOnly && observationPair.Satisfied {
			continue
		}
		controlIds := strings.Join(controlObservationMap[id], ", ")
		rows = append(rows, []string{
			controlIds,
			observationPair.Name,
			convertSatisfied(observationPair.Satisfied, observationPair.StateChange),
			string(observationPair.StateChange),
			observationPair.Observation,
			observationPair.ComparedObservation,
		})
	}
	var err error
	if len(rows) != 0 {
		err = message.Table(header, rows, columnSize)
	}

	return noObservations, err
}

// NewResultComparisonMap -> create a map of result comparisons from two OSCAL results
func NewResultComparisonMap(result oscalTypes.Result, comparedResult oscalTypes.Result) map[string]ResultComparison {
	findingMap := generateFindingMap(*result.Findings)
	comparedFindingMap := generateFindingMap(*comparedResult.Findings)

	relatedObservationsMap := make(map[string][]*oscalTypes.Observation)
	comparedRelatedObservationsMap := make(map[string][]*oscalTypes.Observation)
	resultComparisonMap := make(map[string]ResultComparison)

	if result.Observations != nil {
		relatedObservationsMap = generateRelatedObservationsMap(findingMap, generateObservationMap(*result.Observations))
	}
	if comparedResult.Observations != nil {
		comparedRelatedObservationsMap = generateRelatedObservationsMap(comparedFindingMap, generateObservationMap(*comparedResult.Observations))
	}

	for targetId, finding := range findingMap {
		comparedFinding, found := comparedFindingMap[targetId]
		if !found {
			// Capture new findings that were not found in the compared findings
			resultComparisonMap[targetId] = newResultComparison(finding, nil, relatedObservationsMap[targetId], nil)
		} else {
			// Both findings exist, compare them
			resultComparisonMap[targetId] = newResultComparison(finding, comparedFinding, relatedObservationsMap[targetId], comparedRelatedObservationsMap[targetId])
		}
	}

	for targetId, comparedFinding := range comparedFindingMap {
		_, found := findingMap[targetId]
		if !found {
			// Capture compared findings that were removed/missing from result
			resultComparisonMap[targetId] = newResultComparison(nil, comparedFinding, nil, comparedRelatedObservationsMap[targetId])
		}
	}

	return resultComparisonMap
}

// GetResultComparisonMap gets the result comparison category from the result comparison map
func GetResultComparisonMap(resultComparisonMap map[string]ResultComparison, stateChange StateChange, satisfied bool) ResultComparisonMap {
	ResultComparisonMap := make(ResultComparisonMap)
	for targetId, resultComparison := range resultComparisonMap {
		if resultComparison.StateChange == stateChange && resultComparison.Satisfied == satisfied {
			ResultComparisonMap[targetId] = resultComparison
		}
	}
	return ResultComparisonMap
}

// Collapse map[string]ResultComparisonMap to single ResultComparisonMap
// ** Note this function assumes all unique entities in each ResultComparisonMap
func Collapse(mapResultComparisonMap map[string]ResultComparisonMap) ResultComparisonMap {
	resultComparisonMap := make(ResultComparisonMap)
	for _, v := range mapResultComparisonMap {
		for k, v := range v {
			resultComparisonMap[k] = v
		}
	}
	return resultComparisonMap
}

// Refactor observations by controls
func RefactorObservationsByControls(ResultComparisonMap ResultComparisonMap) (map[string]ObservationPair, map[string][]string, []string) {
	// for each category, add the ObservationPair and add controlId
	observationPairMap := make(map[string]ObservationPair)
	controlObservationMap := make(map[string][]string)
	noObservations := make([]string, 0)

	for targetId, r := range ResultComparisonMap {
		for _, o := range r.ObservationPairs {
			observationPairMap[o.Name] = *o
			controlObservationMap[o.Name] = append(controlObservationMap[o.Name], targetId)
		}
		if len(r.ObservationPairs) == 0 {
			noObservations = append(noObservations, targetId)
		}
	}

	return observationPairMap, controlObservationMap, noObservations
}

// GetMachineFriendlyObservations returns a machine-readable output of diagnosable observations (e.g., SATISFIED_TO_NOT_SATISFIED)
func GetMachineFriendlyObservations(resultComparisonMap ResultComparisonMap) map[StateChange]interface{} {
	observations := make(map[StateChange]interface{})

	for _, resultComparison := range resultComparisonMap {
		if resultComparison.ObservationPairs != nil {
			for _, op := range resultComparison.ObservationPairs {
				if _, ok := observations[op.StateChange]; !ok {
					observations[op.StateChange] = make([]any, 0)
				}
				observations[op.StateChange] = append(observations[op.StateChange].([]any), map[string]string{
					"new_observation":      op.ObservationUuid,
					"original_observation": op.ComparedObservationUuid,
				})
			}
		}
	}

	return observations
}

// newResultComparison create new result comparison from two findings
func newResultComparison(finding *oscalTypes.Finding, comparedFinding *oscalTypes.Finding, relatedObservations []*oscalTypes.Observation, comparedRelatedObservations []*oscalTypes.Observation) ResultComparison {
	var state StateChange
	var satisfied bool
	observationPairs := CreateObservationPairs(relatedObservations, comparedRelatedObservations)

	if comparedFinding == nil {
		state = NEW
		satisfied = finding.Target.Status.State == "satisfied"
	} else if finding == nil {
		state = REMOVED
	} else {
		state = compareFindings(finding, comparedFinding)
		satisfied = finding.Target.Status.State == "satisfied"
	}

	resultComparison := ResultComparison{
		StateChange:      state,
		Satisfied:        satisfied,
		Finding:          finding,
		ComparedFinding:  comparedFinding,
		ObservationPairs: observationPairs,
	}
	return resultComparison
}

// generateFindingMap creates a finding map on the TargetId
// ** Note: this assumes 1:1 relationship between targetId and finding
func generateFindingMap(findings []oscalTypes.Finding) map[string]*oscalTypes.Finding {
	findingMap := make(map[string]*oscalTypes.Finding, len(findings))
	for i := range findings {
		finding := &findings[i]
		findingMap[finding.Target.TargetId] = finding
	}
	return findingMap
}

// generateObservationMap creates observations map on a slice of observations
func generateObservationMap(observations []oscalTypes.Observation) map[string]*oscalTypes.Observation {
	observationMap := make(map[string]*oscalTypes.Observation, len(observations))

	for i := range observations {
		observation := &observations[i]
		observationMap[observation.UUID] = observation
	}

	return observationMap
}

// generateRelatedObservationsMap creates observations map on the TargetId from the findingMap and observationMap
// ** Note: this assumes 1:1 relationship between targetId and finding
func generateRelatedObservationsMap(findingMap map[string]*oscalTypes.Finding, observationMap map[string]*oscalTypes.Observation) map[string][]*oscalTypes.Observation {
	relatedObservationsMap := make(map[string][]*oscalTypes.Observation, len(findingMap))

	for i := range findingMap {
		relatedObservations := findingMap[i].RelatedObservations
		observations := make([]*oscalTypes.Observation, 0)
		if relatedObservations != nil {
			for _, relatedObservation := range *relatedObservations {
				if observation, ok := observationMap[relatedObservation.ObservationUuid]; ok {
					if observation != nil {
						observations = append(observations, observation)
					}
				}
			}
		}
		relatedObservationsMap[i] = observations
	}

	return relatedObservationsMap
}

// compareFindings compares the target.status.state of two findings and calculates the state change between the two
func compareFindings(finding *oscalTypes.Finding, comparedFinding *oscalTypes.Finding) StateChange {
	var state StateChange = UNCHANGED

	if finding == nil {
		if comparedFinding != nil {
			state = REMOVED
		}
	} else {
		if comparedFinding == nil {
			state = NEW
		} else {
			status := finding.Target.Status.State
			comparedStatus := comparedFinding.Target.Status.State

			if status == "not-satisfied" && comparedStatus == "satisfied" {
				state = SATISFIED_TO_NOT_SATISFIED
			} else if status == "satisfied" && comparedStatus == "not-satisfied" {
				state = NOT_SATISFIED_TO_SATISFIED
			}
		}
	}

	return state
}

// convertSatisfied converts the satisfied boolean to a string
func convertSatisfied(satisfied bool, stateChange StateChange) string {
	if stateChange == REMOVED {
		return "N/A"
	} else {
		return strconv.FormatBool(satisfied)
	}
}
