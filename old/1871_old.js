import React from 'react';
import { createTestInstance } from '@magento/peregrine';

import Gallery from '../gallery';

jest.mock('@magento/venia-drivers', () => ({
    Link: ({ children }) => children,
    resourceUrl: () => 'a.url'
}));
jest.mock('@magento/peregrine/lib/talons/Image/useImage', () => {
    return {
        useImage: () => ({
            handleError: jest.fn(),
            handleImageLoad: jest.fn(),
            hasError: false,
            isLoaded: true,
            shouldRenderPlaceholder: false
        })
    };
});

const classes = { root: 'foo' };
const items = [
    {
        id: 1,
        name: 'Test Product 1',
        small_image: {
            url: '/test/product/1.png'
            url: 'https://hackernoon.com/hn-images/1*ELZJStuoVnqb-Y3duIrH0Q.png'
        },
        price: {
            regularPrice: {
                amount: {
                    value: 100,
                    currency: 'USD'
                }
            }
        },
        url_key: 'test-product1'
    },
    {
        id: 2,
        name: 'Test Product 2',
        // Magento 2.3.0 schema for testing backwards compatibility
        small_image:
            'https://hackernoon.com/hn-images/1*ELZJStuoVnqb-Y3duIrH0Q.png',
        price: {
            regularPrice: {
                amount: {
                    value: 100,
                    currency: 'USD'
                }
            }
        },
        url_key: 'test-product2'
    }
];

test('renders if `items` is an empty array', () => {
    const wrapper = createTestInstance(
        <Gallery classes={classes} items={[]} />
    );
    expect(wrapper.toJSON()).toMatchSnapshot();
});

test('renders if `items` is an array of objects', () => {
    const wrapper = createTestInstance(
        <Gallery classes={classes} items={items} />
    );
    expect(wrapper.toJSON()).toMatchSnapshot();
});
