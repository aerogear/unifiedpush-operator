# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

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
