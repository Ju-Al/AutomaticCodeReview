import React, { useMemo } from 'react';
import React, { useCallback, useState } from 'react';

    // TODO: Replace "doneEditing" with a query for existing data.
    const [doneEditing, setDoneEditing] = useState(false);
    const handleClick = useCallback(() => {
        setDoneEditing(true);
        onSave();
    }, [onSave]);

    const className = doneEditing
        ? defaultClasses.container
        : defaultClasses.container_edit_mode;

    /**
     * TODO
     *
     * Change this to reflect diff UI in diff mode.
     */
    const shippingMethod = doneEditing ? (
        <div>In Read Only Mode</div>
    ) : (
        <div>In Edit Mode</div>
    );
    return (
        <div className={className}>
            <div>Shipping Method Will be handled in PWA-179</div>
            <div className={defaultClasses.text_content}>{shippingMethod}</div>
            {!doneEditing ? (
                <div className={defaultClasses.proceed_button_container}>
                    <Button onClick={handleClick} priority="normal">
                        {'Continue to Payment Information'}
                    </Button>
                </div>
            ) : null}
        </div>
import Button from '../../Button';
    displayStates,
    useShippingMethod
} from '@magento/peregrine/lib/talons/CheckoutPage/useShippingMethod';

import { mergeClasses } from '../../../classify';

import Done from './done';
import Editing from './editing';
import defaultClasses from './shippingMethod.css';

import {
    GET_SHIPPING_METHODS,
    GET_SELECTED_SHIPPING_METHOD,
    SET_SHIPPING_METHOD
} from './shippingMethod.gql';

const ShippingMethod = props => {
    const { onSave } = props;

    const talonProps = useShippingMethod({
        onSave,
        queries: {
            getShippingMethods: GET_SHIPPING_METHODS,
            getSelectedShippingMethod: GET_SELECTED_SHIPPING_METHOD
        },
        mutations: {
            setShippingMethod: SET_SHIPPING_METHOD
        }
    });

    const {
        displayState,
        handleSubmit,
        hasShippingMethods,
        isLoadingShippingMethods,
        isLoadingSelectedShippingMethod,
        selectedShippingMethod,
        shippingMethods,
        showEditMode
    } = talonProps;

    const classes = mergeClasses(defaultClasses, props.classes);

    const contentsMap = useMemo(
        () =>
            new Map()
                .set(
                    displayStates.EDITING,
                    <Editing
                        handleSubmit={handleSubmit}
                        hasShippingMethods={hasShippingMethods}
                        isLoading={
                            isLoadingSelectedShippingMethod ||
                            isLoadingShippingMethods
                        }
                        selectedShippingMethod={selectedShippingMethod}
                        shippingMethods={shippingMethods}
                    />
                )
                .set(
                    displayStates.DONE,
                    <Done
                        isLoading={isLoadingSelectedShippingMethod}
                        selectedShippingMethod={selectedShippingMethod}
                        shippingMethods={shippingMethods}
                        showEditMode={showEditMode}
                    />
                ),
        [
            handleSubmit,
            hasShippingMethods,
            isLoadingShippingMethods,
            isLoadingSelectedShippingMethod,
            selectedShippingMethod,
            shippingMethods,
            showEditMode
        ]
    );

    const containerClass =
        displayState === displayStates.EDITING ? classes.root : classes.done;
    const contents = contentsMap.get(displayState);

    return <div className={containerClass}>{contents}</div>;
};

export default ShippingMethod;
