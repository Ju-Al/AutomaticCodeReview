/******************************************************************************
*                 SOFA, Simulation Open-Framework Architecture                *
*                    (c) 2006 INRIA, USTL, UJF, CNRS, MGH                     *
*                                                                             *
* This program is free software; you can redistribute it and/or modify it     *
* under the terms of the GNU Lesser General Public License as published by    *
* the Free Software Foundation; either version 2.1 of the License, or (at     *
* your option) any later version.                                             *
*                                                                             *
* This program is distributed in the hope that it will be useful, but WITHOUT *
* ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or       *
* FITNESS FOR A PARTICULAR PURPOSE. See the GNU Lesser General Public License *
* for more details.                                                           *
*                                                                             *
* You should have received a copy of the GNU Lesser General Public License    *
* along with this program. If not, see <http://www.gnu.org/licenses/>.        *
*******************************************************************************
* Authors: The SOFA Team and external contributors (see Authors.txt)          *
*                                                                             *
* Contact information: contact@sofa-framework.org                             *
******************************************************************************/
#include <SofaMiscSolver/NewmarkImplicitSolver.h>
#include <sofa/core/visual/VisualParams.h>
#include <sofa/simulation/MechanicalVisitor.h>
#include <sofa/simulation/MechanicalOperations.h>
#include <sofa/simulation/VectorOperations.h>
#include <sofa/core/ObjectFactory.h>
#include <cmath>
#include <iostream>
#include <sofa/helper/system/thread/CTime.h>




namespace sofa
{

namespace component
{

namespace odesolver
{
using core::VecId;
using namespace sofa::defaulttype;
using namespace core::behavior;

NewmarkImplicitSolver::NewmarkImplicitSolver()
    : d_rayleighStiffness(initData(&d_rayleighStiffness,0.0,"rayleighStiffness","Rayleigh damping coefficient related to stiffness") )
    , d_rayleighMass( initData(&d_rayleighMass,0.0,"rayleighMass","Rayleigh damping coefficient related to mass"))
    , d_velocityDamping( initData(&d_velocityDamping,0.0,"vdamping","Velocity decay coefficient (no decay if null)") )
    , d_gamma( initData(&d_gamma, 0.5, "gamma", "Newmark scheme gamma coefficient"))
    , d_beta( initData(&d_beta, 0.25, "beta", "Newmark scheme beta coefficient") )
    , d_threadSafeVisitor(initData(&d_threadSafeVisitor, false, "threadSafeVisitor", "If true, do not use realloc and free visitors in fwdInteractionForceField."))
{
    cpt=0;
}


void NewmarkImplicitSolver::solve(const core::ExecParams* params, SReal dt, sofa::core::MultiVecCoordId xResult, sofa::core::MultiVecDerivId vResult)
{
    sofa::simulation::common::VectorOperations vop( params, this->getContext() );
    sofa::simulation::common::MechanicalOperations mop( params, this->getContext() );
    MultiVecCoord pos(&vop, core::VecCoordId::position() );
    MultiVecDeriv vel(&vop, core::VecDerivId::velocity() );
    MultiVecDeriv b(&vop);
    MultiVecDeriv aResult(&vop);
    MultiVecCoord newPos(&vop, xResult );
    MultiVecDeriv newVel(&vop, vResult );


    // dx is no longer allocated by default (but it will be deleted automatically by the mechanical objects)
    MultiVecDeriv dx(&vop, core::VecDerivId::dx()); dx.realloc(&vop, !d_threadSafeVisitor.getValue(), true);
    const double gamma = f_gamma.getValue();
    const double beta = f_beta.getValue();
    const double rM = f_rayleighMass.getValue();
    const double rK = f_rayleighStiffness.getValue();
    const bool verbose  = f_verbose.getValue();

    /* This integration scheme is based on the following equations:
    *
    *   $x_{t+h} = x_t + h v_t + h^2/2 ( (1-2\beta) a_t + 2\beta a_{t+h} )$
    *   $v_{t+h} = v_t + h ( (1-\gamma) a_t + \gamma a_{t+h} )$
    *
    * Applied to a mechanical system where $ M a_t + (r_M M + r_K K) v_t + K x_t
    = f_ext$, we need to solve the following system:
    *
    *   $ M a_{t+h} + (r_M M + r_K K) v_{t+h} + K x_{t+h} = f_ext $
    *   $ M a_{t+h} + (r_M M + r_K K) ( v_t + h ( (1-\gamma) a_t + \gamma a_{t+h} ) ) + K ( x_t + h v_t + h^2/2 ( (1-2\beta) a_t + 2\beta a_{t+h} ) ) =
    f_ext $
    *   $ ( M + h \gamma (r_M M + r_K K) + h^2 \beta K ) a_{t+h} = f_ext - (r_M
    M + r_K K) ( v_t + h (1-\gamma) a_t ) - K ( x_t + h v_t + h^2/2 (1-2\beta) a_t )
    $
    *   $ ( (1 + h \gamma r_M) M + (h^2 \beta + h \gamma r_K) K ) a_{t+h} =
    f_ext - (r_M M + r_K K) v_t - K x_t - (r_M M + r_K K) ( h (1-\gamma) a_t ) - K (
    h v_t + h^2/2 (1-2\beta) a_t) $
    *   $ ( (1 + h \gamma r_M) M + (h^2 \beta + h \gamma r_K) K ) a_{t+h} = Ma_t
    - (r_M M + r_K K) ( h (1-\gamma) a_t ) - K ( h v_t + h^2/2 (1-2\beta) a_t ) $
    *
    * The current implementation first computes $a_t$ directly (as in the
    explicit solvers), then solves the previous system to compute $a_{t+dt}$, and
    finally computes the new position and velocity.
    */

    //we need to initialize a_t and to store it as a vecId to be used in the resolution of this solver (using as well old xand v). Once we have a_{t+dt} we

    if(cpt ==0)
    if( verbose )
     {
        msg_info() << "NewmarkImplicitSolver, aPrevious = " << a;
        msg_info() << "NewmarkImplicitSolver, xPrevious = " << pos;
        msg_info() << "NewmarkImplicitSolver, vPrevious = " << vel;
    }
//can update the new x and v.
    const SReal h = dt;
    const double gamma = d_gamma.getValue();
    const double beta = d_beta.getValue();
    const double rM = d_rayleighMass.getValue();
    const double rK = d_rayleighStiffness.getValue();

    // 1. Initialize a_t and to store it as a vecId to be used in the resolution of this solver (using as well old xand v)
    // Once we have a_{t+dt} we can update the new x and v.

    if (cpt == 0 || this->getContext()->getTime()==0.0)
    {
        vop.v_alloc(pID);
    }

    // Define a
    MultiVecDeriv a(&vop, pID);
    a.realloc(&vop, !d_threadSafeVisitor.getValue(), true);

    if(cpt == 0)
    {
        a.clear();
        mop.computeAcc(0,a,pos,vel);
    }
    cpt++;

    msg_info() << "NewmarkImplicitSolver, aPrevious = " << a;
    msg_info() << "NewmarkImplicitSolver, xPrevious = " << pos;
    msg_info() << "NewmarkImplicitSolver, vPrevious = " << vel;


    // 2. Compute right hand term of equation on a_{t+h}

    mop.computeForce(b,true,false);
    //b = f;
    // b = M a
    if (rM != 0.0 || rK != 0.0 || beta != 0.5)
    {
        mop.propagateDx(a);

        // b += (-h (1-\gamma)(r_M M + r_K K) - h^2/2 (1-2\beta) K ) a
        mop.addMBKdx(b, -h*(1-gamma)*rM, h*(1-gamma), h*(1-gamma)*rK + h*h*(1-2*beta)/2.0,true,true);
    }

    // b += -h K v
    mop.addMBKv(b, -rM, 1,rK+h);

    msg_info() << "NewmarkImplicitSolver, b = " << b;
    mop.projectResponse(b);          // b is projected to the constrained space
    msg_info() << "NewmarkImplicitSolver, projected b = " << b;


    // 3. Solve system of equations on a_{t+h}

    core::behavior::MultiMatrix<simulation::common::MechanicalOperations> matrix(&mop);

    matrix = MechanicalMatrix::K * (-h*h*beta - h*rK*gamma) + MechanicalMatrix::B*(-h)*gamma + MechanicalMatrix::M * (1 + h*gamma*rM);

    msg_info()<<"NewmarkImplicitSolver, matrix = "<< MechanicalMatrix::K *(-h*h*beta + -h*rK*gamma) + MechanicalMatrix::M * (1 + h*gamma*rM) << " = " << matrix<<sendl;
    matrix.solve(aResult, b);
    msg_info() << "NewmarkImplicitSolver, a1 = " << aResult;


    // 4. Compute the new position and velocity

#ifdef SOFA_NO_VMULTIOP // unoptimized version
    // x_{t+h} = x_t + h v_t + h^2/2 ( (1-2\beta) a_t + 2\beta a_{t+h} )
    b.eq(vel, a, h*(0.5-beta));
    b.peq(aResult, h*beta);
    newPos.eq(pos, b, h);
    solveConstraint(dt,xResult,core::ConstraintParams::POS);
    // v_{t+h} = v_t + h ( (1-\gamma) a_t + \gamma a_{t+h} )
    newVel.eq(vel, a, h*(1-gamma));
    newVel.peq(aResult, h*gamma);
    solveConstraint(dt,vResult,core::ConstraintParams::VEL);

#else // single-operation optimization
    typedef core::behavior::BaseMechanicalState::VMultiOp VMultiOp;

    VMultiOp ops;
    ops.resize(2);
    ops[0].first = newPos;
    ops[0].second.push_back(std::make_pair(pos.id(),1.0));
    ops[0].second.push_back(std::make_pair(vel.id(), h));
    ops[0].second.push_back(std::make_pair(a.id(), h*h*(0.5-beta)));
    ops[0].second.push_back(std::make_pair(aResult.id(),h*h*beta));//b=vt+at*h/2(1-2*beta)+a(t+h)*h*beta
    ops[1].first = newVel;
    ops[1].second.push_back(std::make_pair(vel.id(),1.0));
    ops[1].second.push_back(std::make_pair(a.id(), h*(1-gamma)));
    ops[1].second.push_back(std::make_pair(aResult.id(),h*gamma));//v(t+h)=vt+at*h*(1-gamma)+a(t+h)*h*gamma
    vop.v_multiop(ops);

    mop.solveConstraint(vResult,core::ConstraintParams::VEL);
    mop.solveConstraint(xResult,core::ConstraintParams::POS);

#endif

    mop.addSeparateGravity(dt, newVel);	// v += dt*g . Used if mass wants to add G separately from the other forces to v.
    if (d_velocityDamping.getValue()!=0.0)
        newVel *= exp(-h*d_velocityDamping.getValue());


    msg_info() << "NewmarkImplicitSolver, final x = " << newPos;
    msg_info() << "NewmarkImplicitSolver, final v = " << newVel;

    // Increment
    a.eq(aResult);
}

int NewmarkImplicitSolverClass = core::RegisterObject("Implicit time integratorusing Newmark scheme")
        .add< NewmarkImplicitSolver >();

} // namespace odesolver

} // namespace component

} // namespace sofa

