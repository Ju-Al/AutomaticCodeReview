/*
Copyright 2020 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ingress

import "errors"

// ErrNotFound is the error when a broker doesn't exist in the configmap.
// This can happen if the clients specifies invalid broker in the path, or the configmap volume hasn't been updated.
var ErrNotFound = errors.New("not found")

// ErrIncomplete is the error when a broker entry exists in the configmap but its decouple queue is nil or empty.
// This should never happen unless there is a bug in the controller.
var ErrIncomplete = errors.New("incomplete config")

// ErrNotReady is the error when a broker is not ready.
var ErrNotReady = errors.New("not ready")
// ErrOverflow is the error when a broker ingress is sending too many requests to PubSub.
var ErrOverflow = errors.New("bundler reached buffered byte limit")
