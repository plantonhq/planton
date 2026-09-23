# GcpVertexAiTensorboard Guide

The judgment this guide protects: a TensorBoard is shared history. Make one per team and region, let training jobs fill it, and declare only the experiments and runs something else depends on -- then protect it, because a destroy takes every series ever logged.

## Jobs attach by name

A Vertex AI custom training job, a pipeline step, or the Vertex AI SDK (`aiplatform.init(experiment=..., experiment_tensorboard=...)`) attaches to a TensorBoard by its full resource name -- the `name` output. The job also needs a service account allowed to write to it and a Cloud Storage staging bucket; the TensorBoard itself stores the logged series in Google's tenant storage (the `blob_storage_path_prefix` output shows where).

## Experiments and runs: declared or discovered

Training code creates experiments and runs as it logs, and most teams never declare one. Declare an experiment when something must find it before the first job runs -- a dashboard, an evaluation pipeline, a naming convention the team enforces -- and declare a run when it is a fixed point of comparison, like a baseline. Declared and job-created children live side by side: the block never touches the runs jobs create, and jobs can write into declared runs by id.

Ids are 1-128 lowercase letters, digits, and hyphens, and fixed once created. Google requires every run in an experiment to have a display name unique within that experiment; a run's display name defaults to its id, and the spec rejects a collision before Google sees it.

## Identity comes from Google

Google assigns the TensorBoard a numeric id at creation. The `name` output carries it inside the full resource name; `tensorboard_id` is the id alone. The declared experiments and runs address the TensorBoard by that id on both engines, so they are created after it and deleted before it.

## Encryption and destroy

`kmsKeyName` encrypts the TensorBoard and every logged series under your key; the Vertex AI Service Agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it, and the key is fixed at creation. `deletionPolicy` fans to every declared experiment and run: `DELETE` removes the TensorBoard and every series in it -- the job-created ones included -- `PREVENT` makes destroy fail, and `ABANDON` leaves everything in place. A TensorBoard a team compares against should carry `PREVENT`.
