import actions from './actions';

        if (history) {
            const filters = { ...chosenFilterOptions, [group]: newState };
            updateCatalogUrl(filters, history);
        }
    };

export const filterRemove = ({ group, title, value }, history) =>
    async function thunk(dispatch, getState) {
        const {
            catalog: { chosenFilterOptions }
        } = getState();

        const oldState = chosenFilterOptions[group] || [];
        const newState = oldState.filter(item => {
            return item.title !== title || item.value !== value;
        });

        dispatch(actions.filterOption.remove({ newState, group }));

        if (history) {
            const filters = { ...chosenFilterOptions, [group]: newState };
            updateCatalogUrl(filters, history);
        }
    };

export const getAllCategories = () =>
    async function thunk(dispatch) {
        dispatch(actions.getAllCategories.request());

        try {
            // TODO: implement rest or graphql call for categories
            // `/rest/V1/categories` requires auth for some reason
            // TODO: we need to configure Jest to support dynamic imports
            // const { default: payload } = await import('./mockData');

            dispatch(actions.getAllCategories.receive(mockData));
        } catch (error) {
            dispatch(actions.getAllCategories.receive(error));
        }
    };

export const setCurrentPage = payload =>
    async function thunk(dispatch) {
        dispatch(actions.setCurrentPage.receive(payload));
    };

export const setPrevPageTotal = payload =>
    async function thunk(dispatch) {
        dispatch(actions.setPrevPageTotal.receive(payload));
    };
