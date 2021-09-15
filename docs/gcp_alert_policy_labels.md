# GCP Alert Policy Labels
> ## How Revere understands Google Cloud Monitoring Alert Policies

Revere parses incoming Cloud Monitoring Alert incidents based on the "user labels" associated with the Alert Policy itself.

This avoids using inconsistent naming schemes or the added complexity of attempting to parse alert conditions themselves.

Label keys and values are subject to [Google's format and content restrictions](https://cloud.google.com/resource-manager/docs/creating-managing-labels#requirements).

| Key | Meaning | Schema | Example |
|:---:|:-------:|:------:|:-------:|
| `revere-service-name` | "What is the 'short name' of the service?" | Arbitrary string, read based on Revere's config file | `buffer` |
| `revere-service-environment` | "Where does this instance of the service operate?" | Arbitrary string, read based on Revere's config file | `prod` |
| `revere-alert-type` | "What does this alert firing mean" | One of `degraded-performance`, `partial-outage`, or `major-outage` | `major-outage` |

The label values do not need to be unique: multiple alert policies can have the same labels to be understood the same way by Revere.

