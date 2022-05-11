# ibcmapper

This package is separated into multiple version directories because different versions of `cosmos/ibc-go` cannot
be imported into the same project. Having separate versions that align with the version of `cosmos/ibc-go` allows
more control over imported packages.

## Usage

Add this line to the require portion of your go.mod and update the version or path as needed:
```
github.com/figment-networks/ni-cosmoslib/ibcmapper/v2 v2.0.0
```

## Creating a Release

Adjust the patch version as needed in the commands below:

To create a v2 release:
```
git tag ibcmapper/v2.0.0
```

To create a v1 release:
```
git tag ibcmapper/v1.0.0
```
