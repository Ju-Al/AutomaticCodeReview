import React, { useMemo } from 'react';
import React, { Component } from 'react';
import { func, number, shape, string } from 'prop-types';

import classify from '../../classify';
import { mergeClasses } from '../../classify';
import defaultClasses from './pagination.css';
import Tile from './tile';
import NavButton from './navButton';
import { navButtons } from './constants';

const tileBuffer = 2;

class Pagination extends Component {
    static propTypes = {
        classes: shape({
            root: string
        }),
        pageControl: shape({
            currentPage: number,
            setPage: func,
            totalPages: number
        })
    };

    get navigationTiles() {
        const { pageControl } = this.props;
        const { currentPage, setPage, totalPages } = pageControl;

        // Begin building page navigation tiles
        const tiles = [];
        const visibleBuffer = Math.min(tileBuffer * 2, totalPages - 1);
        const leadTile = this.getLeadTile(currentPage, totalPages);

        for (let i = leadTile; i <= leadTile + visibleBuffer; i++) {
            const tile = i;
            tiles.push(tile);
        }
        // End building page navigation tiles

        return tiles.map(tileNumber => {
            return (
                <Tile
                    isActive={tileNumber === currentPage}
                    key={tileNumber}
                    number={tileNumber}
                    onClick={setPage}
                />
            );
        });
    }

    render() {
        const { classes } = this.props;
        const { currentPage, totalPages } = this.props.pageControl;

        if (!this.props.pageControl || totalPages == 1) {
            return null;
        }

        const { navigationTiles } = this;
        const isActiveLeft = !(currentPage == 1);
        const isActiveRight = !(currentPage == totalPages);

        return (
            <div className={classes.root}>
                <NavButton
                    name={navButtons.firstPage.name}
                    active={isActiveLeft}
                    onClick={this.leftSkip}
                    buttonLabel={navButtons.firstPage.buttonLabel}
                />
                <NavButton
                    name={navButtons.prevPage.name}
                    active={isActiveLeft}
                    onClick={this.slideNavLeft}
                    buttonLabel={navButtons.prevPage.buttonLabel}
                />
                {navigationTiles}
                <NavButton
                    name={navButtons.nextPage.name}
                    active={isActiveRight}
                    onClick={this.slideNavRight}
                    buttonLabel={navButtons.nextPage.buttonLabel}
                />
                <NavButton
                    name={navButtons.lastPage.name}
                    active={isActiveRight}
                    onClick={this.rightSkip}
                    buttonLabel={navButtons.lastPage.buttonLabel}
                />
            </div>
        );
import { usePagination } from '../../../../peregrine/lib/talons/Pagination/usePagination';

const Pagination = props => {
    const { currentPage, setPage, totalPages } = props.pageControl;

    const talonProps = usePagination({
        currentPage,
        setPage,
        totalPages
    });

    const {
        handleLeftSkip,
        handleRightSkip,
        handleNavBack,
        handleNavForward,
        isActiveLeft,
        isActiveRight,
        tiles
    } = talonProps;

    const navigationTiles = useMemo(
        () =>
            tiles.map(tileNumber => {
                return (
                    <Tile
                        isActive={tileNumber === currentPage}
                        key={tileNumber}
                        number={tileNumber}
                        onClick={setPage}
                    />
                );
            }),
        [currentPage, tiles, setPage]
    );

    if (totalPages == 1) {
        return null;
    }

    const classes = mergeClasses(defaultClasses, props.classes);

    return (
        <div className={classes.root}>
            <NavButton
                name={navButtons.firstPage.name}
                active={isActiveLeft}
                onClick={handleLeftSkip}
                buttonLabel={navButtons.firstPage.buttonLabel}
            />
            <NavButton
                name={navButtons.prevPage.name}
                active={isActiveLeft}
                onClick={handleNavBack}
                buttonLabel={navButtons.prevPage.buttonLabel}
            />
            {navigationTiles}
            <NavButton
                name={navButtons.nextPage.name}
                active={isActiveRight}
                onClick={handleNavForward}
                buttonLabel={navButtons.nextPage.buttonLabel}
            />
            <NavButton
                name={navButtons.lastPage.name}
                active={isActiveRight}
                onClick={handleRightSkip}
                buttonLabel={navButtons.lastPage.buttonLabel}
            />
        </div>
    );
};

Pagination.propTypes = {
    classes: shape({
        root: string
    }),
    pageControl: shape({
        currentPage: number,
        setPage: func,
        totalPages: number
    }).isRequired
};

export default Pagination;
