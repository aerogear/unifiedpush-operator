# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [0.5.1] - 2021-8-9
### Changed
- Alerts updated to support Openshift 4.8.

## [0.5.0] - 2020-01-12
### Added
- New field databaseSecret to UnifiedPushServer CRD spec.

### Removed
- Logic that cleans up DeploymentConfigs & ImageStreams after migration

### Changed
- Reduced permissions given to operator SA, since it no longer uses DC/IS

## [0.4.2] - 2019-12-12
### Fixed
- Re-add POSTGRES_VERSION to secret to allow backup jobs to succeed again

## [0.4.1] - 2019-12-11
### Fixed
- UnifiedPushDatabaseDown alert always fired if using external database

## [0.4.0] - 2019-12-03
### Added
- UnifiedpushServer CRs can contain coordinates for an external PostgreSQL database
- UnifiedpushServer CR status block has more useful information

### Changed
- Operator SA permissions are limited to the namespace now, no cluster permissions are given
- Operator creates the Monitoring Resources (GrafanaDashboard, PrometheusRule and ServiceMonitor) on creation of the UnifiedPushServer CR.
- Operator creates its own Monitoring Resources (GrafanaDashboard, PrometheusRule) on installation.

### Removed
- This version of the operator no longer deals with the PushApplication or *Variant CRDs

## [0.3.0] - 2019-10-10
### Changed
- Use Deployments instead of DeploymentConfigs
- Use static image references instead of ImageStreams

### Removed
- Delete old resources that were created by the previous versions of the operator
  (DeploymentConfig and ImageStreams)

## [0.2.0] - 2019-09-05
### Added
- `PushApplication`, `AndroidVariant`, and `IOSVariant` all now store
  their relevant IDs (from UPS) as annotations in their CRs.

### Changed
- Some values received from UPS (notably `MasterSecret` for
  `PushApplication`, `Secret` for "variant" kinds), will now be
  updated in the CR Status if they're "renewed" throught the UPS Admin
  UI. This update may take some time (next sync interval, since
  there's no event to trigger immediate reconciliation).

### Deprecated
- The ID fields in the Status block (`PushApplicationId`, `VariantId`)
  currently only exist for compatibility, and are likely to be removed
  in a future version.

## [0.1.2] - 2019-08-21
### Changed
- Added resource limits for UPS Server, oauthproxy, and postgres pods
- Fixed an issue where enmasse resources would be cached in a bad state

## [0.1.1] - 2019-08-06
### Changed
- Made delete methods in UPS client package idempotent
- Added step to the SOP to re-deploy UPS & PostgreSQL

## [0.1.0] - 2019-07-24
### Added
- Initial implementation of the unifiedpush-operator
- `UnifiedPushServer` kind to deploy a UPS deployment (only one supported)
- `PushApplication` kind to create an application in your UPS deployment
- `AndroidVariant` kind to create an Android variant for an application
- `IOSVariant` kind to create an iOS variant for an application
