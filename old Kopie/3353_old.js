import { useCallback } from 'react';

export const SHIMMER_TYPE_SUFFIX = '_SHIMMER';

export default rootType => {
    const [
        ,
        {
            actions: { setNextRootComponent }
        }
    ] = useAppContext();
    const type = `${rootType.toUpperCase()}${SHIMMER_TYPE_SUFFIX}`;

    const setShimmerType = useCallback(() => {
        setNextRootComponent(type);
    }, [setNextRootComponent, type]);

    return {
        setShimmerType
    };
};
