/**
 * Copyright (c) 2013-present, Facebook, Inc.
 *
 * @emails react-core
 * @flow
 */

// $FlowExpectedError
import navCommunity from '../../content/community/nav.yml';
// $FlowExpectedError
import navDocs from '../../content/docs/nav.yml';
// $FlowExpectedError
import navTutorial from '../../content/tutorial/nav.yml';
// $FlowExpectedError
import navFooter from '../../content/footerNav.yml';
// $FlowExpectedError
import navHeader from '../../content/headerNav.yml';

const sectionListDocs = navDocs.map(
  (item: Object): Object => ({
    ...item,
    directory: 'docs',
  }),
);

const sectionListCommunity = navCommunity.map(
  (item: Object): Object => ({
    ...item,
    directory: 'community',
  }),
);

export {
  sectionListCommunity,
  sectionListDocs,
  navTutorial as sectionListTutorial,
  navHeader as sectionListHeader,
};
