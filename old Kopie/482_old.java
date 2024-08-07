package openfoodfacts.github.scrachx.openfood.fragments;

import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.graphics.Bitmap;
import android.os.AsyncTask;
import android.os.Bundle;
import android.support.annotation.Nullable;
import android.util.Log;
import android.view.LayoutInflater;
import android.view.View;
import android.view.ViewGroup;
import android.widget.Button;
import android.widget.ListView;
import android.widget.Toast;

import com.afollestad.materialdialogs.MaterialDialog;

import java.io.File;
import java.util.ArrayList;
import java.util.List;

import butterknife.BindView;
import butterknife.OnClick;
import butterknife.OnItemClick;
import butterknife.OnItemLongClick;
import openfoodfacts.github.scrachx.openfood.R;
import openfoodfacts.github.scrachx.openfood.models.ProductImageField;
import openfoodfacts.github.scrachx.openfood.models.SaveItem;
import openfoodfacts.github.scrachx.openfood.models.SendProduct;
import openfoodfacts.github.scrachx.openfood.models.SendProductDao;
import openfoodfacts.github.scrachx.openfood.network.OpenFoodAPIClient;
import openfoodfacts.github.scrachx.openfood.utils.Utils;
import openfoodfacts.github.scrachx.openfood.views.SaveProductOfflineActivity;
import openfoodfacts.github.scrachx.openfood.views.adapters.SaveListAdapter;

import static org.apache.commons.lang3.StringUtils.isEmpty;
import static org.apache.commons.lang3.StringUtils.isNotEmpty;

public class OfflineEditFragment extends BaseFragment {

    public static final String LOG_TAG = "OFFLINE_EDIT";
    @BindView(R.id.listOfflineSave) ListView listView;
    @BindView(R.id.buttonSendAll) Button buttonSend;
    private List<SaveItem> saveItems;
    private String loginS, passS;

    @Override
    public View onCreateView(LayoutInflater inflater, ViewGroup container, Bundle savedInstanceState) {
        return createView(inflater, container, R.layout.fragment_offline_edit);
    }

    @Override
    public void onViewCreated(View view, @Nullable Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);

        final SharedPreferences settingsLogin = getContext().getSharedPreferences("login", 0);
        final SharedPreferences settingsUsage = getContext().getSharedPreferences("usage", 0);
        saveItems = new ArrayList<>();
        loginS = settingsLogin.getString("user", "");
        passS = settingsLogin.getString("pass", "");
        boolean firstUse = settingsUsage.getBoolean("firstOffline", false);
        if(!firstUse) {
            new MaterialDialog.Builder(getContext())
                    .title(R.string.title_info_dialog)
                    .content(R.string.text_offline_info_dialog)
                    .onPositive((dialog, which) -> {
                        SharedPreferences.Editor editor = settingsUsage.edit();
                        editor.putBoolean("firstOffline", true);
                        editor.apply();
                    })
                    .positiveText(R.string.txtOk)
                    .show();
        }
        buttonSend.setEnabled(false);
    }

    @OnItemClick(R.id.listOfflineSave)
    protected void OnClickListOffline(int position) {
        Intent intent = new Intent(getActivity(), SaveProductOfflineActivity.class);
        SaveItem si = (SaveItem) listView.getItemAtPosition(position);
        intent.putExtra("barcode", si.getBarcode());
        startActivity(intent);
    }

    @OnItemLongClick(R.id.listOfflineSave)
    protected boolean OnLongClickListOffline(int position) {
        final int lapos = position;
        new MaterialDialog.Builder(getActivity())
                .title(R.string.txtDialogsTitle)
                .content(R.string.txtDialogsContentDelete)
                .positiveText(R.string.txtYes)
                .negativeText(R.string.txtNo)
                .onPositive((dialog, which) -> {
                    String barcode = saveItems.get(lapos).getBarcode();
                    SendProduct.deleteAll(SendProduct.class, "barcode = ?", barcode);
                    final SaveListAdapter sl = (SaveListAdapter) listView.getAdapter();
                    saveItems.remove(lapos);
                    getActivity().runOnUiThread(() -> sl.notifyDataSetChanged());
                })
                .show();
        return true;
    }

    @OnClick(R.id.buttonSendAll)
    protected void onSendAllProducts() {
        new MaterialDialog.Builder(getActivity())
                .title(R.string.txtDialogsTitle)
                .content(R.string.txtDialogsContentSend)
                .positiveText(R.string.txtYes)
                .negativeText(R.string.txtNo)
                .onPositive((dialog, which) -> {
                    OpenFoodAPIClient apiClient = new OpenFoodAPIClient(getContext());
                    final List<SendProduct> listSaveProduct = Utils.getAppDaoSession(getActivity()).getSendProductDao().loadAll();
                    for (final SendProduct product : listSaveProduct) {
                        if (isEmpty(product.getBarcode()) || isEmpty(product.getImgupload_front())) {
                            continue;
                        }

                        if(!loginS.isEmpty() && !passS.isEmpty()) {
                            product.setUserId(loginS);
                            product.setPassword(passS);
                        }

                        if(isNotEmpty(product.getImgupload_ingredients())) {
                            product.compress(ProductImageField.INGREDIENTS);
                        }

                        if(isNotEmpty(product.getImgupload_nutrition())) {
                            product.compress(ProductImageField.NUTRITION);
                        }

                        if(isNotEmpty(product.getImgupload_front())) {
                            product.compress(ProductImageField.FRONT);
                        }

                        apiClient.post(getActivity(), product, value -> {
                            if (value) {
                                int productIndex = listSaveProduct.indexOf(product);

                                if (productIndex >= 0 && productIndex < saveItems.size()) {
                                    saveItems.remove(productIndex);
                                }

                                ((SaveListAdapter) listView.getAdapter()).notifyDataSetChanged();
                                Utils.getAppDaoSession(getActivity()).getSendProductDao().deleteInTx(Utils.getAppDaoSession(getActivity()).getSendProductDao().queryBuilder().where(SendProductDao.Properties.Barcode.eq(product.getBarcode())).list());
                            }
                        });
                    }
                })
                .show();
    }

    @Override
    public void onResume() {
        super.onResume();
        new FillAdapter().execute(getActivity());
    }

    public class FillAdapter extends AsyncTask<Context, Void, Context> {

        @Override
        protected void onPreExecute() {
            saveItems.clear();
            List<SendProduct> listSaveProduct = Utils.getAppDaoSession(getActivity()).getSendProductDao().loadAll();
            if (listSaveProduct.size() == 0) {
                Toast.makeText(getActivity(), R.string.txtNoData, Toast.LENGTH_LONG).show();
            } else {
                Toast.makeText(getActivity(), R.string.txtLoading, Toast.LENGTH_LONG).show();
            }
        }

        @Override
        protected Context doInBackground(Context... ctx) {
            List<SendProduct> listSaveProduct = Utils.getAppDaoSession(getActivity()).getSendProductDao().loadAll();

            int imageIcon = R.drawable.ic_ok;

            for (SendProduct product : listSaveProduct) {
                if (isEmpty(product.getBarcode()) || isEmpty(product.getImgupload_front())
                        || isEmpty(product.getBrands()) || isEmpty(product.getWeight()) || isEmpty(product.getName())) {
                    imageIcon = R.drawable.ic_no;
                }

                Bitmap bitmap = Utils.decodeFile(new File(product.getImgupload_front()));
                if (bitmap == null) {
                    Log.e(LOG_TAG, "Unable to load the image of the product: " + product.getBarcode());
                    continue;
                }

                Bitmap imgUrl = Bitmap.createScaledBitmap(bitmap, 200, 200, true);
                saveItems.add(new SaveItem(product.getName(), imageIcon, imgUrl, product.getBarcode()));
            }

            return ctx[0];
        }

        @Override
        protected void onPostExecute(Context ctx) {
            List<SendProduct> listSaveProduct = Utils.getAppDaoSession(getActivity()).getSendProductDao().loadAll();
            if (listSaveProduct.isEmpty()) {
                return;
            }

            SaveListAdapter adapter = new SaveListAdapter(ctx, saveItems);
            listView.setAdapter(adapter);

            boolean canSend = true;
            for (SendProduct sp : listSaveProduct) {
                if (isEmpty(sp.getBarcode()) || isEmpty(sp.getImgupload_front())) {
                    canSend = false;
                    break;
                }
            }

            buttonSend.setEnabled(canSend);
        }
    }
}
