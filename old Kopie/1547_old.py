# -*- Mode: python; tab-width: 4; indent-tabs-mode:nil; coding:utf-8 -*-
# vim: tabstop=4 expandtab shiftwidth=4 softtabstop=4 fileencoding=utf-8
#
# MDAnalysis --- http://www.mdanalysis.org
# Copyright (c) 2006-2017 The MDAnalysis Development Team and contributors
# (see the file AUTHORS for the full list of names)
#
# Released under the GNU Public Licence, v2 or any higher version
#
# Please cite your use of MDAnalysis in published work:
#
# R. J. Gowers, M. Linke, J. Barnoud, T. J. E. Reddy, M. N. Melo, S. L. Seyler,
# D. L. Dotson, J. Domanski, S. Buchoux, I. M. Kenney, and O. Beckstein.
# MDAnalysis: A Python package for the rapid analysis of molecular dynamics
# simulations. In S. Benthall and S. Rostrup editors, Proceedings of the 15th
# Python in Science Conference, pages 102-109, Austin, TX, 2016. SciPy.
#
# N. Michaud-Agrawal, E. J. Denning, T. B. Woolf, and O. Beckstein.
# MDAnalysis: A Toolkit for the Analysis of Molecular Dynamics Simulations.
# J. Comput. Chem. 32 (2011), 2319--2327, doi:10.1002/jcc.21787
#
from __future__ import division, absolute_import

import pytest
from six.moves import range

import numpy as np

from numpy.testing import assert_equal, assert_array_equal

import MDAnalysis as mda
from MDAnalysis.analysis import base

from MDAnalysisTests.datafiles import PSF, DCD
from MDAnalysisTests.util import no_deprecated_call


class FrameAnalysis(base.AnalysisBase):
    """Just grabs frame numbers of frames it goes over"""
    def __init__(self, reader, **kwargs):
        super(FrameAnalysis, self).__init__(reader, **kwargs)
        self.traj = reader
        self.frames = []

    def _single_frame(self):
        self.frames.append(self._ts.frame)


class IncompleteAnalysis(base.AnalysisBase):
    def __init__(self, reader, **kwargs):
        super(IncompleteAnalysis, self).__init__(reader, **kwargs)


class OldAPIAnalysis(base.AnalysisBase):
    """for version 0.15.0"""
    def __init__(self, reader, **kwargs):
        self._setup_frames(reader, **kwargs)

    def _single_frame(self):
        pass


class TestAnalysisBase(object):
    @staticmethod
    @pytest.fixture()
    def u():
        return mda.Universe(PSF, DCD)

    def test_default(self, u):
        an = FrameAnalysis(u.trajectory).run()
        assert an.n_frames == len(u.trajectory)
        assert_equal(an.frames, list(range(len(u.trajectory))))

    def test_start(self, u):
        an = FrameAnalysis(u.trajectory, start=20).run()
        assert an.n_frames == len(u.trajectory) - 20
        assert_equal(an.frames, list(range(20, len(u.trajectory))))

    def test_stop(self, u):
        an = FrameAnalysis(u.trajectory, stop=20).run()
        assert an.n_frames == 20
        assert_equal(an.frames, list(range(20)))

    def test_step(self, u):
        an = FrameAnalysis(u.trajectory, step=20).run()
        assert an.n_frames == 5
        assert_equal(an.frames, list(range(98))[::20])

    def test_verbose(self, u):
        a = FrameAnalysis(u.trajectory, verbose=True)
        assert a._verbose
        assert not a._quiet

    def test_incomplete_defined_analysis(self, u):
        with pytest.raises(NotImplementedError):
            IncompleteAnalysis(u.trajectory).run()

    def test_old_api(self, u):
        OldAPIAnalysis(u.trajectory).run()

    def test_start_stop_step_conversion(self, u):
        an = FrameAnalysis(u.trajectory)
        assert an.start == 0
        assert an.stop == u.trajectory.n_frames
        assert an.step == 1


def test_filter_baseanalysis_kwargs():
    def bad_f(mobile, step=2):
        pass

    def good_f(mobile, ref):
        pass

    kwargs = {'step': 3, 'foo': None}

    with pytest.raises(ValueError):
        base._filter_baseanalysis_kwargs(bad_f, kwargs)

    base_kwargs, kwargs = base._filter_baseanalysis_kwargs(good_f, kwargs)

    assert 1 == len(kwargs)
    assert kwargs['foo'] == None

    assert 5 == len(base_kwargs)
    assert base_kwargs['start'] is None
    assert base_kwargs['step'] == 3
    assert base_kwargs['stop'] is None
    assert base_kwargs['quiet'] is None
    assert base_kwargs['verbose'] is None


def simple_function(mobile):
    return mobile.center_of_geometry()


def test_AnalysisFromFunction():
    u = mda.Universe(PSF, DCD)
    step = 2
    ana1 = base.AnalysisFromFunction(simple_function, mobile=u.atoms,
                                     step=step).run()
    ana2 = base.AnalysisFromFunction(simple_function, u.atoms,
                                     step=step).run()
    ana3 = base.AnalysisFromFunction(simple_function, u.trajectory, u.atoms,
                                     step=step).run()
    ana2 = base.AnalysisFromFunction(simple_function, u.atoms, step=step).run()
    ana3 = base.AnalysisFromFunction(simple_function, u.trajectory, u.atoms, step=step).run()

    results = []
    for ts in u.trajectory[::step]:
        results.append(simple_function(u.atoms))
    results = np.asarray(results)

    for ana in (ana1, ana2, ana3):
        assert_array_equal(results, ana.results)


def test_analysis_class():
    ana_class = base.analysis_class(simple_function)
    assert issubclass(ana_class, base.AnalysisBase)
    assert issubclass(ana_class, base.AnalysisFromFunction)

    u = mda.Universe(PSF, DCD)
    step = 2
    ana = ana_class(u.atoms, step=step).run()

    results = []
    for ts in u.trajectory[::step]:
        results.append(simple_function(u.atoms))
    results = np.asarray(results)

    assert_array_equal(results, ana.results)
    with pytest.raises(ValueError):
        ana_class(2)


def test_analysis_class_decorator():
    # Issue #1511
    # analysis_class should not raise
    # a DeprecationWarning
    u = mda.Universe(PSF, DCD)

    def distance(a, b):
        return np.linalg.norm((a.centroid() - b.centroid()))

    Distances = base.analysis_class(distance)

    with no_deprecated_call():
        d = Distances(u.atoms[:10], u.atoms[10:20]).run()
