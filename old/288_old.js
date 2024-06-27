////////////////////////////////////////////////////////////////////////////
const constants = require('./constants');
const util = require('./util');
module.exports = {
    create,
};

class Results {}
util.createMethods(Results.prototype, constants.objectTypes.RESULTS, [
function create(realmId, info) {
    return util.createList(Results.prototype, realmId, info);
//
// Copyright 2016 Realm Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////

'use strict';

import { objectTypes } from './constants';
import { createList, createMethods } from './util';

export class Results {
    constructor() {
        throw new TypeError('Illegal constructor');
    }
}

createMethods(Results.prototype, objectTypes.RESULTS, [
    'filtered',
    'sorted',
    'snapshot',
]);

export function create(realmId, info) {
    return createList(Results.prototype, realmId, info);
}
