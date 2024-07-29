
        $builder->add('media', CroppableMediaLinkAdminType::class, [
            'required' => true,
        ]);
        $builder->add('caption', TextType::class, array(
            'required' => false
        ));
        $builder->add('altText', TextType::class, array(
            'required' => false,
            'label' => 'mediapagepart.image.alttext'
        ));
        $builder->add('link', URLChooserType::class, array(
            'required' => false,
            'label' => 'mediapagepart.image.link'
        ));
        $builder->add('openInNewWindow', CheckboxType::class, array(
            'required' => false,
            'label' => 'mediapagepart.image.openinnewwindow'
        ));
    }

    /**
     * Returns the name of this type.
     *
     * @return string The name of this type
     */
    public function getBlockPrefix()
    {
        return '{{ pagepart|lower }}type';
    }

    /**
     * Sets the default options for this type.
     *
     * @param OptionsResolver $resolver The resolver for the options.
     */
    public function configureOptions(OptionsResolver $resolver)
    {
        $resolver->setDefaults(array(
            'data_class' => '\{{ namespace }}\Entity\PageParts\{{ pagepart }}',
        ));
    }
}
