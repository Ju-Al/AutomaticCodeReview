    ]);
    return (
      <div className={componentStyles}>
        {isIncentivizedTestnet && (
          <>
            <div className={styles.backgroundOverlay} />
            <SVGInline
              svg={backgroundImage}
              className={styles.backgroundImage}
            />
          </>
        )}
      </div>
    );
  }
}
