# Using the Azure provider extension with Gardener as operator

The [`core.gardener.cloud/v1alpha1.CloudProfile` resource](https://github.com/gardener/gardener/blob/master/example/30-cloudprofile.yaml) declares a `providerConfig` field that is meant to contain provider-specific configuration.

In this document we are describing how this configuration looks like for Azure and provide an example `CloudProfile` manifest with minimal configuration that you can use to allow creating Azure shoot clusters.

## `CloudProfileConfig`

An example `CloudProfileConfig` for the Azure extension looks as follows:

```yaml
apiVersion: azure.provider.extensions.gardener.cloud/v1alpha1
kind: CloudProfileConfig
countUpdateDomains:
- region: westeurope
  count: 5
countFaultDomains:
- region: westeurope
  count: 3
acceleratedNetworingMachineTypes:
- Standard_D3_v2
machineImages:
- name: coreos
  versions:
  - version: 2135.6.0
    urn: "CoreOS:CoreOS:Stable:2135.6.0"
    acceleratedNetworking: true
  - version: 2303.3.0
    id: "/Subscriptions/4bfa08b6-bad8-4b8e-aa00-741c0a859e36/Providers/Microsoft.Compute/Locations/westus/Publishers/CoreOS/ArtifactTypes/VMImage/Offers/CoreOS/Skus/Stable/Versions/2303.3.0"
```

The cloud profile configuration contains information about the update via `.countUpdateDomains[]` and failure domain via `.countFaultDomains[]` counts in the Azure regions you want to offer.

Additionally, it contains the real machine image identifiers in the Azure environment. You have to map every image version that you specify in `.spec.machineImages[].versions` such that the Azure extension knows the machine image identifiers for every version you want to offer. You can provide either URN (`.machineImages[].versions[].urn`) or image ID (`.machineImages[].versions[].id`) for each image version. It's also possible to specify for each image version if the version is supporting Azure accelerated networking via `.machineImages[].versions[].acceleratedNetworking`.

Via `.acceleratedNetworingMachineTypes` you can specify the machine types which come into question for usage of Azure accelerated networking.


## Example `CloudProfile` manifest

The possible values for `.spec.volumeTypes[].name` on Azure are `Standard_LRS`, `StandardSSD_LRS` and `Premium_LRS`. There is another volume type called `UltraSSD_LRS` but this type is not supported to use as os disk. If an end user select a volume type whose name is not equal to one of the valid values then the machine will be created with the default volume type which belong to the selected machine type. Therefore it is recommended to configure only the valid values for the `.spec.volumeType[].name` in the `CloudProfile`.

Please find below an example `CloudProfile` manifest:

```yaml
apiVersion: core.gardener.cloud/v1beta1
kind: CloudProfile
metadata:
  name: azure
spec:
  type: azure
  kubernetes:
    versions:
    - version: 1.16.1
    - version: 1.16.0
      expirationDate: "2020-04-05T01:02:03Z"
  machineImages:
  - name: coreos
    versions:
    - version: 2135.6.0
  machineTypes:
  - name: Standard_D4_v3
    cpu: "4"
    gpu: "0"
    memory: 16Gi
  volumeTypes:
  - name: Standard_LRS
    class: standard
    usable: true
  - name: StandardSSD_LRS
    class: premium
    usable: false
  - name: Premium_LRS
    class: premium
    usable: false
  regions:
  - name: westeurope
  providerConfig:
    apiVersion: azure.provider.extensions.gardener.cloud/v1alpha1
    kind: CloudProfileConfig
    countUpdateDomains:
    - region: westeurope
      count: 5
    countFaultDomains:
    - region: westeurope
      count: 3
    machineImages:
    - name: coreos
      versions:
      - version: 2135.6.0
        urn: "CoreOS:CoreOS:Stable:2135.6.0"
```
