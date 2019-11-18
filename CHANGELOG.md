# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

- Operator creates the Monitoring Resources (GrafanaDashboard, PrometheusRule and ServiceMonitor) on creation of the UnifiedPushServer CR.
- Operator creates its own Monitoring Resources (GrafanaDashboard, PrometheusRule) on installation.
- Removal of the templates for the related Monitoring Resources.

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
