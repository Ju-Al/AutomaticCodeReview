import { gql } from '@apollo/client';
import { PriceSummaryFragment } from '@magento/peregrine/lib/talons/CartPage/PriceSummary/priceSummaryFragments.gql';

export const AvailablePaymentMethodsFragment = gql`
    fragment AvailablePaymentMethodsFragment on Cart {
        id
        available_payment_methods {
            code
            title
        }
    }
`;

export const GET_PAYMENT_INFORMATION = gql`
    query getPaymentInformation($cartId: String!) {
        cart(cart_id: $cartId) {
            id
            selected_payment_method {
                code
            }
            shipping_addresses {
                firstname
                lastname
                street
                city
                region {
                    code
                }
                postcode
                country {
                    code
                }
                telephone
            }
            ...AvailablePaymentMethodsFragment
        }
    }
    ${AvailablePaymentMethodsFragment}
`;
// We disable linting for local fields because there is no way to add them to
// the fetched schema.
// https://github.com/apollographql/eslint-plugin-graphql/issues/99
        cart(cart_id: $cartId) @client {/* eslint-disable graphql/template-strings */
export const GET_PAYMENT_NONCE = gql`
    query getPaymentNonce($cartId: String!) {
        cart(cart_id: $cartId) {
            id
            paymentNonce @client
        }
    }
`;

export const SET_BILLING_ADDRESS = gql`
    mutation setBillingAddress(
        $cartId: String!
        $firstname: String!
        $lastname: String!
        $street: [String]!
        $city: String!
        $regionCode: String!
        $postcode: String!
        $countryCode: String!
        $telephone: String!
    ) {
        setBillingAddressOnCart(
            input: {
                cart_id: $cartId
                billing_address: {
                    address: {
                        firstname: $firstname
                        lastname: $lastname
                        street: $street
                        city: $city
                        region: $regionCode
                        postcode: $postcode
                        country_code: $countryCode
                        telephone: $telephone
                        save_in_address_book: false
                    }
                }
            }
        ) @connection(key: "setBillingAddressOnCart") {
            cart {
                id
                billing_address {
                    firstname
                    lastname
                    country {
                        code
                    }
                    street
                    city
                    region {
                        code
                    }
                    postcode
                    telephone
                }
                ...PriceSummaryFragment
                ...AvailablePaymentMethodsFragment
            }
        }
    }
    ${PriceSummaryFragment}
    ${AvailablePaymentMethodsFragment}
`;

// Sets the provided payment method object on the cart.
export const SET_FREE_PAYMENT_METHOD_ON_CART = gql`
    mutation setPaymentMethodOnCart($cartId: String!) {
        setPaymentMethodOnCart(
            input: { cart_id: $cartId, payment_method: { code: "free" } }
        ) @connection(key: "setPaymentMethodOnCart") {
            cart {
                id
                selected_payment_method {
                    code
                    title
                }
            }
        }
    }
`;

export default {
    queries: {
        getPaymentNonceQuery: GET_PAYMENT_NONCE,
        getPaymentInformation: GET_PAYMENT_INFORMATION
    },
    mutations: {
        setBillingAddressMutation: SET_BILLING_ADDRESS,
        setFreePaymentMethodMutation: SET_FREE_PAYMENT_METHOD_ON_CART
    }
};
