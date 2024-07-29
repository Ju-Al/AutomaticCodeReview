            file.getParentFile().mkdirs();
            PrintWriter out = new PrintWriter(file);
            out.write(template);
            out.close();
        } catch (FileNotFoundException e) {
            e.printStackTrace();
        }
        return true;
    }

    public static boolean generateDashboardForDMNEndpoint(String handlerName, int id, List<String> decisionNames){
        String template = readStandardDashboard();
        template = template.replaceAll("\\$handlerName\\$", handlerName);
        template = template.replaceAll("\\$id\\$", String.valueOf(id));
        template = template.replaceAll("\\$uid\\$", UUID.randomUUID().toString());

        try{
            IJGrafana jgrafana = JGrafana.parse(template);
            decisionNames.forEach(x -> jgrafana.addPanel(PanelType.GRAPH, "Decision " + x, String.format("dmn_result{handler = \"%s\"}", x)));
            File file = new File("/tmp/dashboard-endpoint-" + handlerName + ".json");
            file.getParentFile().mkdirs();
            jgrafana.writeToFile(file);
        } catch (JsonProcessingException e) {
            e.printStackTrace();
        } catch (IOException e) {
            e.printStackTrace();
        }
        return true;
    }
}
