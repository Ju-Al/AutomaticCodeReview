/*
 *      Author: wfg

#include "adios2/ADIOSMPI.h"
#include "adios2/core/Support.h"
#include "adios2/core/adiosFunctions.h"
Engine::Engine(ADIOS &adios, const std::string engineType,
               const std::string &name, const std::string accessMode,
               MPI_Comm mpiComm, const Method &method,
               const std::string endMessage)
: m_MPIComm(mpiComm), m_EngineType(engineType), m_Name(name),
  m_AccessMode(accessMode), m_Method(method), m_ADIOS(adios),
  m_DebugMode(m_Method.m_DebugMode), m_EndMessage(endMessage)
    if (m_DebugMode == true)
    {
        if (m_MPIComm == MPI_COMM_NULL)
        {
            throw std::ios_base::failure(
                "ERROR: engine communicator is MPI_COMM_NULL,"
                " in call to ADIOS Open or Constructor\n");
        }
    }

    MPI_Comm_rank(m_MPIComm, &m_RankMPI);
    MPI_Comm_size(m_MPIComm, &m_SizeMPI);
void Engine::SetCallBack(std::function<void(const void *, std::string,
                                            std::string, std::string, Dims)>
                             callback)
static void EngineThrowUp(const std::string &engineType,
                          const std::string &func)
{
    throw std::invalid_argument(
        "ERROR: Engine bass class " + func + "() called. " + engineType +
        " child class is not implementing this function\n");
}

void Engine::Write(Variable<char> & /*variable*/, const char * /*values*/) {}
void Engine::Write(Variable<unsigned char> & /*variable*/,
                   const unsigned char * /*values*/)
{
}
void Engine::Write(Variable<short> & /*variable*/, const short * /*values*/) {}
void Engine::Write(Variable<unsigned short> & /*variable*/,
                   const unsigned short * /*values*/)
{
}
void Engine::Write(Variable<int> & /*variable*/, const int * /*values*/) {}
void Engine::Write(Variable<unsigned int> & /*variable*/,
                   const unsigned int * /*values*/)
{
}
void Engine::Write(Variable<long int> & /*variable*/,
                   const long int * /*values*/)
{
}
void Engine::Write(Variable<unsigned long int> & /*variable*/,
                   const unsigned long int * /*values*/)
{
}
void Engine::Write(Variable<long long int> & /*variable*/,
                   const long long int * /*values*/)
{
}
void Engine::Write(Variable<unsigned long long int> & /*variable*/,
                   const unsigned long long int * /*values*/)
{
}
void Engine::Write(Variable<float> & /*variable*/, const float * /*values*/) {}
void Engine::Write(Variable<double> & /*variable*/, const double * /*values*/)
{
}
void Engine::Write(Variable<long double> & /*variable*/,
                   const long double * /*values*/)
{
}
void Engine::Write(Variable<std::complex<float>> & /*variable*/,
                   const std::complex<float> * /*values*/)
{
}
void Engine::Write(Variable<std::complex<double>> & /*variable*/,
                   const std::complex<double> * /*values*/)
{
}
void Engine::Write(Variable<std::complex<long double>> & /*variable*/,
                   const std::complex<long double> * /*values*/)
{
}
void Engine::Write(VariableCompound & /*variable*/, const void * /*values*/) {}
void Engine::Write(const std::string & /*variableName*/,
                   const char * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const unsigned char * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const short * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const unsigned short * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/, const int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const unsigned int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const long int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const unsigned long int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const long long int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const unsigned long long int * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const float * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const double * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const long double * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const std::complex<float> * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const std::complex<double> * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const std::complex<long double> * /*values*/)
{
}
void Engine::Write(const std::string & /*variableName*/,
                   const void * /*values*/)
{
}

void Engine::Advance(float /*timeout_sec*/) {}
void Engine::Advance(AdvanceMode /*mode*/, float /*timeout_sec*/) {}
 * Distributed under the OSI-approved Apache License, Version 2.0.  See
 * accompanying file Copyright.txt for details.
 *
 * Engine.cpp
 *
 *  Created on: Dec 19, 2016
 *      Author: William F Godoy godoywf@ornl.gov
 */

#include "Engine.h"
#include "Engine.tcc"

/// \cond EXCLUDE_FROM_DOXYGEN
#include <ios> //std::ios_base::failure
#include <set>
/// \endcond

namespace adios
{

Engine::Engine(const std::string engineType, IO &io, const std::string &name,
               const OpenMode openMode, MPI_Comm mpiComm)
: m_EngineType(engineType), m_IO(io), m_Name(name), m_OpenMode(openMode),
  m_MPIComm(mpiComm), m_DebugMode(io.m_DebugMode)
{
}

void Engine::SetCallBack(
    std::function<void(const void *, std::string, std::string, std::string,
                       std::vector<size_t>)>
        callback)
{
}

// should these functions throw an exception?

void Engine::Advance(const float /*timeout_sec*/) {}
void Engine::Advance(const AdvanceMode /*mode*/, const float /*timeout_sec*/) {}
void Engine::AdvanceAsync(
    AdvanceMode /*mode*/,
    std::function<void(std::shared_ptr<adios::Engine>)> /*callback*/)
{
}

AdvanceStatus Engine::GetAdvanceStatus() { return m_AdvanceStatus; }

void Engine::Close(const int /*transportIndex*/) {}

// READ
void Engine::Release() {}
void Engine::PerformReads(PerformReadMode /*mode*/){};

// PROTECTED
void Engine::Init() {}

void Engine::InitParameters() {}

void Engine::InitTransports() {}

// DoWrite
#define declare_type(T)                                                        \
    void Engine::DoWrite(Variable<T> &variable, const T *values)               \
    {                                                                          \
        ThrowUp("Write");                                                      \
    }
ADIOS2_FOREACH_TYPE_1ARG(declare_type)
#undef declare_type

void Engine::DoWrite(VariableCompound &variable, const void *values)
{ // TODO
}

void Engine::DoWrite(const std::string &variableName, const void *values)
{
    const std::string type(m_IO.GetVariableType(variableName));
    if (m_DebugMode == true)
    {
        if (type.empty() == true)
        {
            throw std::invalid_argument(
                "ERROR: variable " + variableName +
                " was not created with IO.DefineVariable for Engine " + m_Name +
                ", in call to Write\n");
        }
    }

    if (type == "compound")
    {
        DoWrite(m_IO.GetVariableCompound(variableName), values);
    }
#define declare_type(T)                                                        \
    else if (type == GetType<T>())                                             \
    {                                                                          \
        DoWrite(m_IO.GetVariable<T>(variableName),                             \
                reinterpret_cast<const T *>(values));                          \
    }
    ADIOS2_FOREACH_TYPE_1ARG(declare_type)
#undef declare_type
} // end DoWrite

// READ
VariableBase *Engine::InquireVariableUnknown(const std::string &name,
                                             const bool readIn)
{
    return nullptr;
}
Variable<char> *Engine::InquireVariableChar(const std::string &name,
                                            const bool readIn)
{
    return nullptr;
}
Variable<unsigned char> *Engine::InquireVariableUChar(const std::string &name,
                                                      const bool readIn)
{
    return nullptr;
}
Variable<short> *Engine::InquireVariableShort(const std::string &name,
                                              const bool readIn)
{
    return nullptr;
}
Variable<unsigned short> *Engine::InquireVariableUShort(const std::string &name,
                                                        const bool readIn)
{
    return nullptr;
}
Variable<int> *Engine::InquireVariableInt(const std::string &name,
                                          const bool readIn)
{
    return nullptr;
}
Variable<unsigned int> *Engine::InquireVariableUInt(const std::string &name,
                                                    const bool readIn)
{
    return nullptr;
}
Variable<long int> *Engine::InquireVariableLInt(const std::string &name,
                                                const bool readIn)
{
    return nullptr;
}
Variable<unsigned long int> *
Engine::InquireVariableULInt(const std::string &name, const bool readIn)
{
    return nullptr;
}
Variable<long long int> *Engine::InquireVariableLLInt(const std::string &name,
                                                      const bool readIn)
{
    return nullptr;
}
Variable<unsigned long long int> *
Engine::InquireVariableULLInt(const std::string &name, const bool readIn)
{
    return nullptr;
}

Variable<float> *Engine::InquireVariableFloat(const std::string &name,
                                              const bool readIn)
{
    return nullptr;
}
Variable<double> *Engine::InquireVariableDouble(const std::string &name,
                                                const bool readIn)
{
    return nullptr;
}
Variable<long double> *Engine::InquireVariableLDouble(const std::string &name,
                                                      const bool readIn)
{
    return nullptr;
}
Variable<cfloat> *Engine::InquireVariableCFloat(const std::string &name,
                                                const bool readIn)
{
    return nullptr;
}
Variable<cdouble> *Engine::InquireVariableCDouble(const std::string &name,
                                                  const bool readIn)
{
    return nullptr;
}

Variable<cldouble> *Engine::InquireVariableCLDouble(const std::string &name,
                                                    const bool readIn)
{
    return nullptr;
}

#define declare_type(T)                                                        \
    void Engine::DoScheduleRead(Variable<T> &variable, const T *values)        \
    {                                                                          \
        ThrowUp("ScheduleRead");                                               \
    }                                                                          \
                                                                               \
    void Engine::DoScheduleRead(const std::string &variableName,               \
                                const T *values)                               \
    {                                                                          \
        ThrowUp("ScheduleRead");                                               \
    }
ADIOS2_FOREACH_TYPE_1ARG(declare_type)
#undef declare_type

// PRIVATE
void Engine::ThrowUp(const std::string function)
{
    throw std::invalid_argument("ERROR: Engine derived class " + m_EngineType +
                                " doesn't implement function " + function +
                                "\n");
}

} // end namespace adios
