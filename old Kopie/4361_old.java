package fr.free.nrw.commons.upload.depicts;

        presenter.verifyDepictions();
import android.app.Application;
import android.os.Bundle;
import android.view.LayoutInflater;
import android.view.View;
import android.view.View.OnClickListener;
import android.view.ViewGroup;
import android.widget.ImageView;
import android.widget.ProgressBar;
import android.widget.TextView;
import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.recyclerview.widget.LinearLayoutManager;
import androidx.recyclerview.widget.RecyclerView;
import butterknife.BindView;
import butterknife.ButterKnife;
import butterknife.OnClick;
import com.google.android.material.textfield.TextInputEditText;
import com.google.android.material.textfield.TextInputLayout;
import com.jakewharton.rxbinding2.view.RxView;
import com.jakewharton.rxbinding2.widget.RxTextView;
import fr.free.nrw.commons.R;
import fr.free.nrw.commons.upload.UploadBaseFragment;
import fr.free.nrw.commons.upload.structure.depictions.DepictedItem;
import fr.free.nrw.commons.utils.DialogUtil;
import io.reactivex.android.schedulers.AndroidSchedulers;
import io.reactivex.disposables.Disposable;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.TimeUnit;
import javax.inject.Inject;
import kotlin.Unit;
import timber.log.Timber;


/**
 * Fragment for showing depicted items list in Upload activity after media details
 */
public class DepictsFragment extends UploadBaseFragment implements DepictsContract.View {

    @BindView(R.id.depicts_title)
    TextView depictsTitle;
    @BindView(R.id.depicts_search_container)
    TextInputLayout depictsSearchContainer;
    @BindView(R.id.depicts_search)
    TextInputEditText depictsSearch;
    @BindView(R.id.depictsSearchInProgress)
    ProgressBar depictsSearchInProgress;
    @BindView(R.id.depicts_recycler_view)
    RecyclerView depictsRecyclerView;
    @BindView(R.id.tooltip)
    ImageView tooltip;

    @Inject
    DepictsContract.UserActionListener presenter;
    private UploadDepictsAdapter adapter;
    private Disposable subscribe;
    public static ArrayList<DepictedItem>selectedDepictedItem;

    @Nullable
    @Override
    public android.view.View onCreateView(@NonNull LayoutInflater inflater, @Nullable ViewGroup container,
                                          @Nullable Bundle savedInstanceState) {
        return inflater.inflate(R.layout.upload_depicts_fragment, container, false);
    }

    @Override
    public void onViewCreated(@NonNull android.view.View view, @Nullable Bundle savedInstanceState) {
        super.onViewCreated(view, savedInstanceState);
        ButterKnife.bind(this, view);
        selectedDepictedItem=new ArrayList<DepictedItem>();
        init();
        presenter.getDepictedItems().observe(getViewLifecycleOwner(), this::setDepictsList);
    }

    /**
     * Initialize presenter and views
     */
    private void init() {
        depictsTitle.setText(getString(R.string.step_count, callback.getIndexInViewFlipper(this) + 1,
                callback.getTotalNumberOfSteps(), getString(R.string.depicts_step_title)));
        tooltip.setOnClickListener(new OnClickListener() {
            @Override
            public void onClick(View v) {
                DialogUtil.showAlertDialog(getActivity(), getString(R.string.depicts_step_title), getString(R.string.depicts_tooltip), getString(android.R.string.ok), null, true);
            }
        });
        presenter.onAttachView(this);
        initRecyclerView();
        addTextChangeListenerToSearchBox();
    }

    /**
     * Initialise recyclerView and set adapter
     */
    private void initRecyclerView() {
        adapter = new UploadDepictsAdapter(item -> {
            presenter.onDepictItemClicked(item);
            return Unit.INSTANCE;
        });
        depictsRecyclerView.setLayoutManager(new LinearLayoutManager(getContext()));
        depictsRecyclerView.setAdapter(adapter);
    }

    @Override
    public void goToNextScreen() {
        callback.onNextButtonClicked(callback.getIndexInViewFlipper(this));
    }

    @Override
    public void goToPreviousScreen() {
        callback.onPreviousButtonClicked(callback.getIndexInViewFlipper(this));
    }

    @Override
    public void noDepictionSelected() {
        DialogUtil.showAlertDialog(getActivity(),
            getString(R.string.no_depictions_selected),
            getString(R.string.no_depictions_selected_warning_desc),
            getString(R.string.continue_message),
            getString(R.string.cancel),
            this::goToNextScreen,
            null
        );
    }

    @Override
    public void onDestroyView() {
        super.onDestroyView();
        presenter.onDetachView();
        subscribe.dispose();
    }

    @Override
    public void showProgress(boolean shouldShow) {
        depictsSearchInProgress.setVisibility(shouldShow ? View.VISIBLE : View.GONE);
    }

    @Override
    public void showError(Boolean value) {
        if (value) {
            depictsSearchContainer.setError(getString(R.string.no_depiction_found));
        } else {
            depictsSearchContainer.setErrorEnabled(false);
        }
    }

    @Override
    public void setDepictsList(List<DepictedItem> depictedItemList) {
        adapter.setItems(depictedItemList);
    }

    @OnClick(R.id.depicts_next)
    public void onNextButtonClicked() {
        presenter.verifyDepictions(getActivity().getApplication());
    }

    @OnClick(R.id.depicts_previous)
    public void onPreviousButtonClicked() {
        callback.onPreviousButtonClicked(callback.getIndexInViewFlipper(this));
    }

    /**
     * Text change listener for the edit text view of depicts
     */
    private void addTextChangeListenerToSearchBox() {
        subscribe = RxTextView.textChanges(depictsSearch)
                .doOnEach(v -> depictsSearchContainer.setError(null))
                .takeUntil(RxView.detaches(depictsSearch))
                .debounce(500, TimeUnit.MILLISECONDS)
                .observeOn(AndroidSchedulers.mainThread())
                .subscribe(filter -> searchForDepictions(filter.toString()), Timber::e);
    }

    /**
     * Search for depictions for the following query
     *
     * @param query query string
     */
    private void searchForDepictions(final String query) {
        // set  recent depictsItem  of DepictsRoomDotaBase to recyclerview when query is emapty
        if (query.isEmpty()) {
            adapter.clear();
            adapter.setItems(getRecentDepictesItem(getActivity().getApplication()));
            return;
        } else {
            presenter.searchForDepictions(query);
            adapter.clear();
        }
        presenter.searchForDepictions(query);
    }

    /**
     * Get the depictesItem form DepictsRoomdataBase
     */
    List<DepictedItem> getRecentDepictesItem(final Application application) {
        final List<DepictedItem> depictedItemList = new ArrayList<>();
        final handleDepictsDoa modelDepicts = new handleDepictsDoa(application);
        depictedItemList.addAll(selectedDepictedItem);
        for (int i = 0; i < modelDepicts.allDepicts.size(); i++) {
            final DepictedItem depictedItem = modelDepicts.allDepicts.get(i).getItem();
            depictedItem.setSelected(false);
            depictedItemList.add(depictedItem);
        }
        return depictedItemList;
    }
}
