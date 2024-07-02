package openfoodfacts.github.scrachx.openfood.views;
import android.app.SearchManager;
import android.app.SearchableInfo;
import android.content.Context;
import android.os.Bundle;
import android.support.v7.widget.DividerItemDecoration;
import android.support.v7.widget.LinearLayoutManager;
import android.support.v7.widget.RecyclerView;
import android.support.v7.widget.SearchView;
import android.support.v7.widget.Toolbar;
import android.view.Menu;
import android.view.MenuInflater;
import android.view.MenuItem;
import android.view.View;

import java.util.ArrayList;
import java.util.Collections;
import java.util.Comparator;
import java.util.HashSet;
import java.util.List;
import java.util.Locale;
import java.util.Set;

import io.reactivex.Single;
import openfoodfacts.github.scrachx.openfood.FastScroller;
import openfoodfacts.github.scrachx.openfood.R;
import openfoodfacts.github.scrachx.openfood.models.Additive;
import openfoodfacts.github.scrachx.openfood.models.AdditiveName;
import openfoodfacts.github.scrachx.openfood.models.AdditiveNameDao;
import openfoodfacts.github.scrachx.openfood.repositories.IProductRepository;
import openfoodfacts.github.scrachx.openfood.repositories.ProductRepository;
import openfoodfacts.github.scrachx.openfood.utils.SearchType;
import openfoodfacts.github.scrachx.openfood.utils.Utils;
import openfoodfacts.github.scrachx.openfood.views.adapters.AdditivesAdapter;

public class AdditivesExplorer extends BaseActivity implements AdditivesAdapter.ClickListener {


    private RecyclerView recyclerView;
    private AdditiveNameDao additiveNameDao;
    private IProductRepository productRepository;
    private List<AdditiveName> additives;
    private Toolbar toolbar;


    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_additives_explorer);

        recyclerView = findViewById(R.id.additiveRecyclerView);
        productRepository = ProductRepository.getInstance();
        toolbar = findViewById(R.id.toolbar);

        setSupportActionBar(toolbar);
        getSupportActionBar().setDisplayHomeAsUpEnabled(true);
        getSupportActionBar().setTitle("Additives");


        additiveNameDao = Utils.getAppDaoSession(this).getAdditiveNameDao();
        List<AdditiveName> additivesNames = additiveNameDao.loadAll();
        String languageCode = Locale.getDefault().getLanguage();
        additives = new ArrayList<>();
        Set<Single<AdditiveName>> hs = new HashSet<>();
        for (int i = 0; i < additivesNames.size(); i++) {
            if (additivesNames.get(i).getName().startsWith("E")) {
                String tag = additivesNames.get(i).getAdditiveTag();
                hs.add(productRepository.getAdditiveByTagAndLanguageCode(tag, languageCode));
            }
        }

        Set<AdditiveName> additiveSet = new HashSet<>();
        for (Single<AdditiveName> additiveNameSingle : hs) {
            additiveNameSingle.subscribe(
                    additiveSet::add
            );
        }

        additives.addAll(additiveSet);

        Collections.sort(additives, new Comparator<AdditiveName>() {
            @Override
            public int compare(AdditiveName additiveName, AdditiveName t1) {
                String s1 = additiveName.getName().toLowerCase().replace('x', '0').split("(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)")[1];
                String s2 = t1.getName().toLowerCase().replace('x', '0').split("(?<=\\D)(?=\\d)|(?<=\\d)(?=\\D)")[1];
                return Integer.valueOf(s1).compareTo(Integer.valueOf(s2));
            }
        });


        recyclerView.setLayoutManager(new LinearLayoutManager(AdditivesExplorer.this));
        recyclerView.setAdapter(new AdditivesAdapter(additives, AdditivesExplorer.this));
        recyclerView.addItemDecoration(new DividerItemDecoration(AdditivesExplorer.this, DividerItemDecoration.VERTICAL));


    }


    @Override
    public void onclick(int position, String name) {
        ProductBrowsingListActivity.startActivity(AdditivesExplorer.this, name, SearchType.ADDITIVE);
    }


    @Override
    public boolean onCreateOptionsMenu(Menu menu) {
        getMenuInflater().inflate(R.menu.menu_main, menu);

        SearchManager searchManager = (SearchManager) this.getSystemService(Context.SEARCH_SERVICE);
        MenuItem searchItem = menu.findItem(R.id.action_search);

        SearchView searchView = (SearchView) searchItem.getActionView();
        if (searchManager.getSearchableInfo(this.getComponentName()) != null) {

            searchView.setSearchableInfo(searchManager.getSearchableInfo(this.getComponentName()));
            searchView.setOnQueryTextListener(new SearchView.OnQueryTextListener() {
                @Override
                public boolean onQueryTextSubmit(String query) {
                    return false;
                }

                @Override
                public boolean onQueryTextChange(String newText) {

                    List<AdditiveName> additiveNames = new ArrayList<>();

                    for (int i = 0; i < additives.size(); i++) {
                        if (additives.get(i).getName().toLowerCase().split("- ")[1].startsWith(newText.toLowerCase())) {
                            additiveNames.add(additives.get(i));
                        }
                    }

                    recyclerView.setAdapter(new AdditivesAdapter(additiveNames, AdditivesExplorer.this));
                    recyclerView.getAdapter().notifyDataSetChanged();

                    return true;
                }
            });

        }


        return super.onCreateOptionsMenu(menu);
    }
}


