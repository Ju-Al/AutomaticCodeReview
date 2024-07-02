/*
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package org.apache.iceberg.spark;

/**
 * Captures information about the current job
 * which is used for displaying on the UI
 */
public class JobGroupInfo implements AutoCloseable {
  private String groupId;
  private String description;
  private boolean interruptOnCancel;

  public JobGroupInfo(String groupId, String desc, boolean interruptOnCancel) {
    this.groupId = groupId;
    this.description = desc;
    this.interruptOnCancel = interruptOnCancel;
  }

  public String groupId() {
    return groupId;
  }

  public String description() {
    return description;
  }

  public boolean interruptOnCancel() {
    return interruptOnCancel;
  }

  @Override
  public void close() throws Exception {
    // do nothing
  }
}
