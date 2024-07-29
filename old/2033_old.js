/**
 * Copyright (c) 2013-present, Facebook, Inc.
 *
 * @emails react-core
 * @flow
 */

import Helmet from 'react-helmet';
import React from 'react';
import {urlRoot} from 'site-constants';
// $FlowFixMe This is a valid path
import languages from '../../../content/languages.yml';

const defaultDescription = 'A JavaScript library for building user interfaces';

type Props = {
  title: string,
  ogDescription: string,
  canonicalUrl: string,
};

      href={canonicalUrl.replace(
        urlRoot,
        `https://${
          language.code === 'en' ? '' : `${language.code}.`
        }reactjs.org`,
      )}
    />
  ));
};

const TitleAndMetaTags = ({title, ogDescription, canonicalUrl}: Props) => {
  return (
    <Helmet title={title}>
      <meta property="og:title" content={title} />
      <meta property="og:type" content="website" />
      {canonicalUrl && <meta property="og:url" content={canonicalUrl} />}
      <meta property="og:image" content="/logo-og.png" />
      <meta
        property="og:description"
        content={ogDescription || defaultDescription}
      />
      <meta property="fb:app_id" content="623268441017527" />
      {canonicalUrl && <link rel="canonical" href={canonicalUrl} />}
      {canonicalUrl && alternatePages(canonicalUrl)}
    </Helmet>
  );
};

export default TitleAndMetaTags;
