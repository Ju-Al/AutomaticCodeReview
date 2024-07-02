import React, { useCallback } from 'react';
import React, { Component } from 'react';
import PropTypes from 'prop-types';

import fromRenderProp from '../util/fromRenderProp';

class Item extends Component {
    static propTypes = {
        classes: PropTypes.shape({
            root: PropTypes.string
        }),
        hasFocus: PropTypes.bool,
        isSelected: PropTypes.bool,
        item: PropTypes.any.isRequired,
        itemIndex: PropTypes.number.isRequired,
        render: PropTypes.oneOfType([PropTypes.func, PropTypes.string])
            .isRequired

const getChild = item => (isString(item) ? item : null);

const Item = props => {
    const {
        uniqueID: key,
        classes,
        hasFocus,
        isSelected,
        item,
        itemIndex,
        render,
        updateSelection,
        setFocus,
        ...restProps
    } = props;
    const onClick = useCallback(() => updateSelection(key), [
        updateSelection,
        key
    ]);
    const onFocus = useCallback(() => setFocus(key), [setFocus, key]);
    const customProps = {
        classes,
        hasFocus,
        isSelected,
        item,
        itemIndex,
        onClick,
        onFocus
    };
    const Root = fromRenderProp(render, Object.keys(customProps));
    return (
        <Root className={classes.root} {...customProps} {...restProps}>
            {getChild(item)}
        </Root>
    );
};

Item.propTypes = {
    classes: PropTypes.shape({
        root: PropTypes.string
    }),
    hasFocus: PropTypes.bool,
    isSelected: PropTypes.bool,
    item: PropTypes.any.isRequired,
    itemIndex: PropTypes.number.isRequired,
    render: PropTypes.oneOfType([PropTypes.func, PropTypes.string]).isRequired
};

Item.defaultProps = {
    classes: {},
    hasFocus: false,
    isSelected: false,
    render: 'div'
};

export default Item;
