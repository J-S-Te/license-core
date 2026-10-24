package catalog

import (
	"errors"
	"reflect"
	"testing"
)

var fourBusinessApplications = []string{
	"contract_management",
	"customer_and_opportunity",
	"project_management",
	"data_analysis",
}

var sixApplications = []string{
	"contract_management",
	"customer_and_opportunity",
	"customer_portal",
	"project_management",
	"settlement",
	"data_analysis",
}

func consumerCount(t *testing.T, application string) int {
	t.Helper()
	components, err := Components(application)
	if err != nil {
		t.Fatalf("Components(%q): %v", application, err)
	}
	count := 0
	for _, component := range components {
		if component.Role != RoleInfrastructure {
			count++
		}
	}
	return count
}

func TestCatalogRejectsUnknownAndCrossApplicationService(t *testing.T) {
	if _, err := Components("unknown_application"); !errors.Is(err, ErrUnknownApplication) {
		t.Fatalf("unknown application error = %v", err)
	}
	if _, err := Lookup("contract_management", "customer-api"); !errors.Is(err, ErrUnknownComponent) {
		t.Fatalf("cross-application lookup error = %v", err)
	}
	if _, err := Lookup("contract_management", "unknown-service"); !errors.Is(err, ErrUnknownComponent) {
		t.Fatalf("unknown service error = %v", err)
	}
}

func TestCatalogReturnsIndependentCopies(t *testing.T) {
	first, err := Components("contract_management")
	if err != nil {
		t.Fatalf("Components: %v", err)
	}
	if len(first) == 0 {
		t.Fatal("empty catalog")
	}
	first[0].ServiceID = "mutated"
	first[0].Role = "MUTATED"
	second, err := Components("contract_management")
	if err != nil {
		t.Fatalf("Components second call: %v", err)
	}
	if second[0].ServiceID == "mutated" || second[0].Role == "MUTATED" {
		t.Fatal("catalog exposed mutable internal state")
	}
}

func TestCatalogSeparatesBIInfrastructureAndConsumers(t *testing.T) {
	total := 0
	for _, application := range sixApplications {
		total += consumerCount(t, application)
	}
	if total != 21 {
		t.Fatalf("six-application consumer count = %d, want 21", total)
	}
	four := 0
	for _, application := range fourBusinessApplications {
		four += consumerCount(t, application)
	}
	if four != 16 {
		t.Fatalf("four-application consumer count = %d, want 16", four)
	}
	components, err := Components("data_analysis")
	if err != nil {
		t.Fatalf("Components(data_analysis): %v", err)
	}
	var metabase *Component
	for index := range components {
		if components[index].ServiceID == "data-analysis-metabase" {
			metabase = &components[index]
		}
	}
	if metabase == nil || metabase.Role != RoleInfrastructure || metabase.RegistrationProtocol != RegistrationProtocolV2 {
		t.Fatalf("metabase component = %+v, want infrastructure protocol v2", metabase)
	}
	if err := ValidateSelection("data_analysis", []string{"data-analysis-api", "data-analysis-aggregation-worker", "data-analysis-alert-worker"}); err != nil {
		t.Fatalf("consumer selection without metabase rejected: %v", err)
	}
}

func TestCatalogSelectionCannotDropRequiredConsumers(t *testing.T) {
	if err := ValidateSelection("customer_portal", []string{"portal-api", "portal-report-worker"}); err != nil {
		t.Fatalf("conditional worker may be omitted: %v", err)
	}
	if err := ValidateSelection("customer_portal", []string{"portal-api"}); err == nil {
		t.Fatal("required portal-report-worker was dropped")
	}
	if err := ValidateSelection("contract_management", []string{"contract-api", "contract-worker"}); err != nil {
		t.Fatalf("complete contract selection rejected: %v", err)
	}
	if err := ValidateSelection("contract_management", []string{"contract-api"}); err == nil {
		t.Fatal("required contract-worker was dropped")
	}
	if err := ValidateSelection("contract_management", []string{"contract-api", "contract-worker", "customer-api"}); err == nil {
		t.Fatal("cross-application selection was accepted")
	}
	if err := ValidateSelection("contract_management", []string{"contract-api", "contract-worker", "contract-worker"}); err == nil {
		t.Fatal("duplicate selection was accepted")
	}
}

func TestCatalogComponentsUseSupportedProtocolAndMode(t *testing.T) {
	modes := []RegistrationMode{LegacyImageV1, BridgingComponentV2, BuiltinComponentV2}
	if !reflect.DeepEqual(modes, []RegistrationMode{"LEGACY_IMAGE_V1", "BRIDGING_COMPONENT_V2", "BUILTIN_COMPONENT_V2"}) {
		t.Fatalf("registration modes = %v", modes)
	}
	if RegistrationProtocolV1 != 1 || RegistrationProtocolV2 != 2 || Version != 1 {
		t.Fatalf("protocol/version constants = %d/%d/%d", RegistrationProtocolV1, RegistrationProtocolV2, Version)
	}
	for _, application := range sixApplications {
		components, err := Components(application)
		if err != nil {
			t.Fatalf("Components(%q): %v", application, err)
		}
		for _, component := range components {
			if component.Application != application {
				t.Fatalf("component application mismatch: %+v", component)
			}
			if component.ServiceID == "" || component.Role == "" || component.RegistrationProtocol != RegistrationProtocolV2 {
				t.Fatalf("invalid component: %+v", component)
			}
		}
	}
}
