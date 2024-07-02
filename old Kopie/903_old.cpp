////////////////////////////////////////////////////////////////////////////////
optimizer::optimizer(lbann_comm *comm, DataType learning_rate)
  : m_comm(comm),
    m_weights(nullptr),
    m_learning_rate(learning_rate),
    m_gradient(nullptr),
    m_gradient_staging(nullptr),
    m_gradient_allreduce_needed(false),
    m_gradient_allreduce_started(false),
    m_gradient_allreduce_finished(false) {}
    m_gradient(other.m_gradient),
    m_gradient_staging(other.m_gradient_staging),
    m_gradient_allreduce_needed(other.m_gradient_allreduce_needed),
    m_gradient_allreduce_started(other.m_gradient_allreduce_started),
    m_gradient_allreduce_finished(other.m_gradient_allreduce_finished),
    m_step_time(other.m_step_time)
{
  if (m_gradient != nullptr) {
    m_gradient = m_gradient->Copy();
  }
  if (m_gradient_staging != nullptr) {
    m_gradient_staging = m_gradient_staging->Copy();
  m_gradient_allreduce_needed = other.m_gradient_allreduce_needed;
  m_gradient_allreduce_started = other.m_gradient_allreduce_started;
  m_gradient_allreduce_finished = other.m_gradient_allreduce_finished;
  m_gradient_allreduce_started = other.m_gradient_allreduce_started;

  // Deep copy matrices
  if (m_gradient != nullptr) { delete m_gradient; }
  if (m_gradient_staging != nullptr) { delete m_gradient_staging; }
  m_gradient = other.m_gradient;
  m_gradient_staging = other.m_gradient_staging;
  if (m_gradient != nullptr) {
    m_gradient = m_gradient->Copy();
  if (m_gradient_staging != nullptr) {
    m_gradient_staging = m_gradient_staging->Copy();
  }

optimizer::~optimizer() {
  if (m_gradient != nullptr) { delete m_gradient; }
  if (m_gradient_staging != nullptr)  { delete m_gradient_staging; }
}

  if (!is_initialized()) {
    LBANN_ERROR("attempted to access the weights being optimized before they are set");
const AbsDistMat& optimizer::get_gradient() {
  // Check if gradient is initialized
  if (!is_initialized()) {
    LBANN_ERROR("attempted to access gradients before they are set up");
  // Perform allreduce on staging matrix if needed
  if (m_gradient_allreduce_needed && !m_gradient_allreduce_started) {
    start_gradient_staging_allreduce();
  }
  if (m_gradient_allreduce_started && !m_gradient_allreduce_finished) {
    m_comm->wait(m_gradient_allreduce_req);
    m_gradient_allreduce_finished = true;
  if (m_gradient_allreduce_needed) {
    add_to_gradient(*m_gradient_staging);
  m_gradient_allreduce_needed = false;
  m_gradient_allreduce_started = false;
  m_gradient_allreduce_finished = false;
void optimizer::start_gradient_staging_allreduce() {
  if (!m_gradient_allreduce_needed || m_gradient_allreduce_started) {
    return;
  }

  m_gradient_allreduce_started = true;
  m_comm->nb_allreduce(*m_gradient_staging,
                       m_gradient_staging->RedundantComm(),
                       m_gradient_allreduce_req,
                       El::mpi::SUM);
  m_gradient_allreduce_finished = false;
}

void optimizer::clear_gradient() {

  // Clear matrices
  El::Zero(*m_gradient);

  // Reset gradient allreduce flags
  m_gradient_allreduce_needed = false;
  m_gradient_allreduce_started = false;
  m_gradient_allreduce_finished = false;

}

                                DataType scale) {
  if (!is_initialized()) {
    LBANN_ERROR("attempted to access gradients before they are set up");
// Copyright (c) 2014-2016, Lawrence Livermore National Security, LLC.
// Produced at the Lawrence Livermore National Laboratory.
// Written by the LBANN Research Team (B. Van Essen, et al.) listed in
// the CONTRIBUTORS file. <lbann-dev@llnl.gov>
//
// LLNL-CODE-697807.
// All rights reserved.
//
// This file is part of LBANN: Livermore Big Artificial Neural Network
// Toolkit. For details, see http://software.llnl.gov/LBANN or
// https://github.com/LLNL/LBANN.
//
// Licensed under the Apache License, Version 2.0 (the "Licensee"); you
// may not use this file except in compliance with the License.  You may
// obtain a copy of the License at:
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the license.
////////////////////////////////////////////////////////////////////////////////

#include "lbann/optimizers/optimizer.hpp"
#include "lbann/utils/timer.hpp"

namespace lbann {

std::string to_string(optimizer_gradient_status status) {
  switch (status) {
  case optimizer_gradient_status::ready:
    return "ready";
  case optimizer_gradient_status::cleared:
    return "cleared";
  case optimizer_gradient_status::allreduce_needed:
    return "allreduce needed";
  case optimizer_gradient_status::allreduce_started:
    return "allreduce started";
  default:
    return "unknown";
  }
}

optimizer::optimizer(lbann_comm* comm, DataType learning_rate)
  : m_comm(comm), m_learning_rate(learning_rate) {
  if (m_comm == nullptr) {
    LBANN_ERROR("got null pointer for lbann_comm");
  }
}

optimizer::optimizer(const optimizer& other)
  : m_comm(other.m_comm),
    m_weights(other.m_weights),
    m_gradient(other.m_gradient ? other.m_gradient->Copy() : nullptr),
    m_gradient_v(other.m_gradient_v ? other.m_gradient_v->Copy() : nullptr),
    m_gradient_sources(other.m_gradient_sources),
    m_gradient_status(other.m_gradient_status),
    m_learning_rate(other.m_learning_rate),
    m_step_time(other.m_step_time) {
  if (m_gradient_status == optimizer_gradient_status::allreduce_started) {
    LBANN_ERROR("attempted to copy optimizer while a "
                "gradient allreduce is in progress");
  }
}

optimizer& optimizer::operator=(const optimizer& other) {
  m_comm = other.m_comm;
  m_weights = other.m_weights;
  m_gradient.reset(other.m_gradient ? other.m_gradient->Copy() : nullptr);
  m_gradient_v.reset(other.m_gradient_v ? other.m_gradient_v->Copy() : nullptr);
  m_gradient_sources = other.m_gradient_sources;
  m_gradient_status = other.m_gradient_status;
  m_learning_rate = other.m_learning_rate;
  m_step_time = other.m_step_time;
  if (m_gradient_status == optimizer_gradient_status::allreduce_started) {
    LBANN_ERROR("attempted to copy optimizer while a "
                "gradient allreduce is in progress");
  }
  return *this;
}

description optimizer::get_description() const {
  description desc(get_type() + " optimizer");
  desc.add("Learning rate", m_learning_rate);
  return desc;
}

weights& optimizer::get_weights() {
  // Item 3, p. 23 in "Effective C++", 3rd ed., by Scott Meyers
  return const_cast<weights&>(static_cast<const optimizer&>(*this).get_weights());
}

const weights& optimizer::get_weights() const {
  if (m_weights == nullptr) {
    LBANN_ERROR("attempted to access the weights being optimized "
                "before they are set");
  }
  return *m_weights;
}

AbsDistMat& optimizer::get_gradient() {

  // Make sure gradient matrix has been setup
  if (m_gradient == nullptr) {
    LBANN_ERROR("attempted to access gradient before it is set up");
  }

  // Make sure gradient values are ready
  start_gradient_allreduce();
  finish_gradient_allreduce();
  if (m_gradient_status == optimizer_gradient_status::cleared) {
    El::Zero(*m_gradient);
    m_gradient_status = optimizer_gradient_status::ready;
  }
  if (m_gradient_status != optimizer_gradient_status::ready) {
    std::ostringstream err;
    err << "unexpected gradient status (expected "
        << "\"" << to_string(optimizer_gradient_status::ready) << "\", "
        << "but found \"" << to_string(m_gradient_status) << "\")";
    LBANN_ERROR(err.str());
  }

  // Return gradient
  return *m_gradient;

}

void optimizer::add_to_gradient(const AbsDistMat& gradient,
                                DataType scale,
                                bool allreduce_needed) {

  // Check that matrices have been setup
  if (m_gradient == nullptr || m_gradient_v == nullptr) {
    LBANN_ERROR("attempted to access gradient before it is set up");
  }
  if (scale == DataType(0)) { return; }

  // Make a view or copy of input matrix in correct distribution
  m_gradient_v->Empty();
  m_gradient_v->AlignWith(*m_gradient);
  if (m_gradient_v->DistData() == gradient.DistData()) {
    El::LockedView(*m_gradient_v, gradient);
  } else if (allreduce_needed) {
    std::unique_ptr<AbsDistMat> temp(gradient.Copy());
    get_comm().allreduce(*temp, temp->RedundantComm());
    El::Copy(*temp, *m_gradient_v);
    allreduce_needed = false;
  } else {
    El::Copy(gradient, *m_gradient_v);
  }

  // Add to gradient
  if (m_gradient_status == optimizer_gradient_status::allreduce_started) {
    finish_gradient_allreduce();
  }
  switch (m_gradient_status) {
  case optimizer_gradient_status::ready:
    if (allreduce_needed) {
      El::Scale(DataType(1) / m_gradient->RedundantSize(), *m_gradient);
      m_gradient_status = optimizer_gradient_status::allreduce_needed;
    }
    El::Axpy(scale, *m_gradient_v, *m_gradient);
    break;
  case optimizer_gradient_status::cleared:
    El::Copy(*m_gradient_v, *m_gradient);
    El::Scale(scale, *m_gradient);
    m_gradient_status = (allreduce_needed ?
                         optimizer_gradient_status::allreduce_needed :
                         optimizer_gradient_status::ready);
    break;
  case optimizer_gradient_status::allreduce_needed:
    {
      const auto& scale_ = (allreduce_needed ?
                            scale :
                            scale / m_gradient->RedundantSize());
      El::Axpy(scale_, *m_gradient_v, *m_gradient);
    }
    break;
  case optimizer_gradient_status::allreduce_started:
  default:
    LBANN_ERROR("unexpected gradient status "
                "(" + to_string(m_gradient_status) + ")");
  }

  // Clean up
  m_gradient_v->Empty();

}

void optimizer::clear_gradient() {
  if (m_gradient_status == optimizer_gradient_status::allreduce_started) {
    finish_gradient_allreduce();
  }
  m_gradient_status = optimizer_gradient_status::cleared;
  m_gradient_sources.clear();
}

El::Int optimizer::get_num_gradient_sources() const {
  return m_gradient_sources.size();
}

void optimizer::add_gradient_source(const void* source) {
  if (source != nullptr) {
    m_gradient_sources.insert(source);
  }
}

void optimizer::remove_gradient_source(const void* source) {
  m_gradient_sources.erase(nullptr);
  m_gradient_sources.erase(source);
  if (m_gradient_sources.empty()) {
    start_gradient_allreduce();
  }
}

void optimizer::setup(weights* w) {
  clear_gradient();

  // Set weights being optimized
  if (w != nullptr) { set_weights(w); }
  if (m_weights == nullptr) {
    LBANN_ERROR("attempted to setup optimizer without weights");
  }

  // Initialize matrices
  const auto& height = m_weights->get_matrix_height();
  const auto& width = m_weights->get_matrix_width();
  const AbsDistMat& values = m_weights->get_values();
  m_gradient.reset(AbsDistMat::Instantiate(values.DistData()));
  m_gradient->AlignWith(values);
  m_gradient->Resize(height, width);
  m_gradient_v.reset(AbsDistMat::Instantiate(values.DistData()));
  m_gradient_v->AlignWith(values);
#ifdef HYDROGEN_HAVE_CUB
  if (m_gradient_v->GetLocalDevice() == El::Device::GPU) {
    m_gradient_v->Matrix().SetMemoryMode(1); // CUB GPU memory pool
  }
#endif // HYDROGEN_HAVE_CUB

}

void optimizer::step() {
  if (m_weights == nullptr) {
    LBANN_ERROR("attempted to perform optimization step without weights");
  }
  const auto start_time = get_time();
  step_compute(m_weights->get_values(), get_gradient());
  m_step_time += get_time() - start_time;
}

DataType optimizer::get_learning_rate() const {
  return m_learning_rate;
}

void optimizer::set_learning_rate(DataType learning_rate) {
  m_learning_rate = learning_rate;
};

void optimizer::start_gradient_allreduce() {
  switch (m_gradient_status) {
  case optimizer_gradient_status::allreduce_needed:
    get_comm().nb_allreduce(*m_gradient,
                            m_gradient->RedundantComm(),
                            m_gradient_allreduce_req);
    m_gradient_status = optimizer_gradient_status::allreduce_started;
    break;
  case optimizer_gradient_status::ready:
  case optimizer_gradient_status::cleared:
  case optimizer_gradient_status::allreduce_started:
    break;
  default: LBANN_ERROR("unexpected gradient status "
                       "(" + to_string(m_gradient_status) + ")");
  }
}

void optimizer::finish_gradient_allreduce() {
  switch (m_gradient_status) {
  case optimizer_gradient_status::allreduce_started:
    get_comm().wait(m_gradient_allreduce_req);
    m_gradient_status = optimizer_gradient_status::ready;
    break;
  case optimizer_gradient_status::ready:
  case optimizer_gradient_status::cleared:
    break;
  case optimizer_gradient_status::allreduce_needed:
    LBANN_ERROR("attempted to finish gradient allreduce "
                "before starting it");
    break;
  default:
    LBANN_ERROR("unexpected gradient status "
                "(" + to_string(m_gradient_status) + ")");
  }
}

// =============================
// Checkpointing
// =============================

bool optimizer::save_to_checkpoint_shared(persist& p, std::string m_name) {
  //  m_learning_rate;
  p.write_datatype(persist_type::train, "learning_rate", m_learning_rate);
  return true;
}

bool optimizer::load_from_checkpoint_shared(persist& p, std::string m_name) {
  p.read_datatype(persist_type::train, "learning_rate", &m_learning_rate);
  get_comm().trainer_broadcast(0, m_learning_rate);
  return true;
}

bool optimizer::save_to_checkpoint_distributed(persist& p, std::string m_name) {
  p.write_datatype(persist_type::train, "learning_rate", m_learning_rate);
  return true;
}

bool optimizer::load_from_checkpoint_distributed(persist& p, std::string m_name) {
  p.read_datatype(persist_type::train, "learning_rate", &m_learning_rate);
  return true;
}

} // namespace lbann
