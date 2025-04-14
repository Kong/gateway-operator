// Code generated by hack/generators/testcases-registration/main.go. DO NOT EDIT.
package integration

func init() {
	addTestsToTestSuite(
		TestAIGatewayCreation,
		TestControlPlaneEssentials,
		TestControlPlaneUpdate,
		TestControlPlaneWatchNamespaces,
		TestControlPlaneWhenNoDataPlane,
		TestDataPlaneBlueGreenHorizontalScaling,
		TestDataPlaneBlueGreenResourcesNotDeletedUntilOwnerIsRemoved,
		TestDataPlaneBlueGreenRollout,
		TestDataPlaneEssentials,
		TestDataPlaneHorizontalScaling,
		TestDataPlanePodDisruptionBudget,
		TestDataPlaneServiceExternalTrafficPolicy,
		TestDataPlaneServiceTypes,
		TestDataPlaneSpecifyingServiceName,
		TestDataPlaneUpdate,
		TestDataPlaneValidation,
		TestDataPlaneVolumeMounts,
		TestGatewayClassCreation,
		TestGatewayClassUpdates,
		TestGatewayConfigurationEssentials,
		TestGatewayDataPlaneNetworkPolicy,
		TestGatewayEssentials,
		TestGatewayMultiple,
		TestGatewayWithMultipleListeners,
		TestHTTPRoute,
		TestHTTPRouteWithTLS,
		TestIngressEssentials,
		TestKongPluginInstallationEssentials,
		TestKonnectEntities,
		TestKonnectExtension,
		TestKonnectExtensionKonnectControlPlaneNotFound,
		TestManualGatewayUpgradesAndDowngrades,
		TestScalingDataPlaneThroughGatewayConfiguration,
	)
}
