        }

    }


    @Override
    public void start(Stage stage) throws IOException {
        
        parseParameters(getParameters());

        FXMLLoader loader
            = new FXMLLoader(getClass().getResource("fxml/designer.fxml"));

        DesignerApp owner = new DesignerApp(stage);
        MainDesignerController mainController = new MainDesignerController(owner);

        NodeInfoPanelController nodeInfoPanelController = new NodeInfoPanelController(owner, mainController);
        XPathPanelController xpathPanelController = new XPathPanelController(owner, mainController);
        SourceEditorController sourceEditorController = new SourceEditorController(owner, mainController);
        EventLogController eventLogController = new EventLogController(owner, mainController);

        loader.setControllerFactory(type -> {
            if (type == MainDesignerController.class) {
                return mainController;
            } else if (type == NodeInfoPanelController.class) {
                return nodeInfoPanelController;
            } else if (type == XPathPanelController.class) {
                return xpathPanelController;
            } else if (type == SourceEditorController.class) {
                return sourceEditorController;
            } else if (type == EventLogController.class) {
                return eventLogController;
            } else {
                // default behavior for controllerFactory:
                try {
                    return type.newInstance();
                } catch (Exception exc) {
                    exc.printStackTrace();
                    throw new RuntimeException(exc); // fatal, just bail...
                }
            }
        });

        stage.setOnCloseRequest(e -> mainController.shutdown());

        Parent root = loader.load();
        Scene scene = new Scene(root);

        stage.setTitle("PMD Rule Designer (v " + PMD.VERSION + ')');
        setIcons(stage);

        stage.setScene(scene);
        stage.show();
    }


    private void setIcons(Stage primaryStage) {
        ObservableList<Image> icons = primaryStage.getIcons();
        final String dirPrefix = "icons/app/";
        List<String> imageNames = Arrays.asList("designer_logo.jpeg");

        // TODO make more icon sizes

        List<Image> images = imageNames.stream()
                                       .map(s -> dirPrefix + s)
                                       .map(s -> getClass().getResourceAsStream(s))
                                       .filter(Objects::nonNull)
                                       .map(Image::new)
                                       .collect(Collectors.toList());

        icons.addAll(images);
    }


    public static void main(String[] args) {
        launch(args);
    }
}