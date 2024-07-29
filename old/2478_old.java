
/**
 * @author Matthias Stöckli
 */
public class HumanTaskVariableNameCmmnXmlConverterTest extends AbstractConverterTest {

    private static final String CMMN_RESOURCE = "org/flowable/test/cmmn/converter/humanTaskTaskVariableName.cmmn";

    @Test
    public void convertXMLToModel() throws Exception {
        CmmnModel cmmnModel = readXMLFile(CMMN_RESOURCE);
        validateModel(cmmnModel);
    }

    @Test
    public void convertModelToXML() throws Exception {
        CmmnModel cmmnModel = readXMLFile(CMMN_RESOURCE);
        CmmnModel parsedModel = exportAndReadXMLFile(cmmnModel);
        validateModel(parsedModel);
    }

    public void validateModel(CmmnModel cmmnModel) {
        assertThat(cmmnModel).isNotNull();

        PlanItemDefinition itemDefinition = cmmnModel.findPlanItemDefinition("task1");

        assertThat(itemDefinition)
                .isInstanceOfSatisfying(HumanTask.class, task -> {
                    assertThat(task.getTaskIdVariableName()).isEqualTo("myTaskId");
                });
    }

}