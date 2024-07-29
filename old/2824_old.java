        }

        Bean<FlowableCmmnServices> flowableCmmnServicesBean = (Bean<FlowableCmmnServices>) beanManager.getBeans(FlowableCmmnServices.class).stream()
                .findAny()
                .orElseThrow(
                        () -> new IllegalStateException("CDI BeanManager cannot find an instance of requested type " + FlowableCmmnServices.class.getName()));
        FlowableCmmnServices services = (FlowableCmmnServices) beanManager
                .getReference(flowableCmmnServicesBean, FlowableCmmnServices.class, beanManager.createCreationalContext(flowableCmmnServicesBean));
        services.setCmmnEngine(cmmnEngine);

        return cmmnEngine;
    }

    public void beforeShutdown(@Observes BeforeShutdown event) {
        if (cmmnEngineLookup != null) {
            cmmnEngineLookup.ungetCmmnEngine();
            cmmnEngineLookup = null;
        }
        LOGGER.info("Shutting down flowable-cmmn-cdi");
    }
}
