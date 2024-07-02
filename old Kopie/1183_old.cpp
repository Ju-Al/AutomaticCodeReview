/*
    caffe_copy(data_count, &data_blob_.cpu_data()[current_row_ * data_count],
               &(*top)[0]->mutable_cpu_data()[i * data_count]);
    caffe_copy(label_data_count,
               &label_blob_.cpu_data()[current_row_ * label_data_count],
               &(*top)[1]->mutable_cpu_data()[i * label_data_count]);
TODO:
- load file in a separate thread ("prefetch")
- can be smarter about the memcpy call instead of doing it row-by-row
  :: use util functions caffe_copy, and Blob->offset()
  :: don't forget to update hdf5_daa_layer.cu accordingly
- add ability to shuffle filenames if flag is set
*/
#include <fstream>  // NOLINT(readability/streams)
#include <string>
#include <vector>

#include "hdf5.h"
#include "hdf5_hl.h"
#include "stdint.h"

#include "caffe/layer.hpp"
#include "caffe/util/io.hpp"
#include "caffe/vision_layers.hpp"

namespace caffe {

template <typename Dtype>
HDF5DataLayer<Dtype>::~HDF5DataLayer<Dtype>() { }

// Load data and label from HDF5 filename into the class property blobs.
template <typename Dtype>
void HDF5DataLayer<Dtype>::LoadHDF5FileData(const char* filename) {
  LOG(INFO) << "Loading HDF5 file" << filename;
  hid_t file_id = H5Fopen(filename, H5F_ACC_RDONLY, H5P_DEFAULT);
  if (file_id < 0) {
    LOG(ERROR) << "Failed opening HDF5 file" << filename;
    return;
  }

  int top_size = this->layer_param_.top_size();
  hdf_blobs_.resize(top_size);

  const int MIN_DATA_DIM = 2;
  const int MAX_DATA_DIM = 4;

  for (int i = 0; i < top_size; ++i) {
      hdf_blobs_[i] = shared_ptr<Blob<Dtype> >(new Blob<Dtype>());
      hdf5_load_nd_dataset(file_id, this->layer_param_.top(i).c_str(),
        MIN_DATA_DIM, MAX_DATA_DIM, hdf_blobs_[i].get());
  }

  herr_t status = H5Fclose(file_id);
  CHECK_GE(status, 0) << "Failed to close HDF5 file " << filename;
  CHECK_GE(hdf_blobs_.size(), 1)
      << "At least one 'top' must be specified for HDF5 data layer";
  int num = hdf_blobs_[0]->num();
  for (int i = 1; i < top_size; ++i) {
      CHECK_EQ(hdf_blobs_[i]->num(), num);
  }
  LOG(INFO) << "Successully loaded " << hdf_blobs_[0]->num() << " rows";
}

template <typename Dtype>
void HDF5DataLayer<Dtype>::LayerSetUp(const vector<Blob<Dtype>*>& bottom,
      vector<Blob<Dtype>*>* top) {
  // Read the source to parse the filenames.
  const string& source = this->layer_param_.hdf5_data_param().source();
  LOG(INFO) << "Loading filename from " << source;
  hdf_filenames_.clear();
  std::ifstream source_file(source.c_str());
  if (source_file.is_open()) {
    std::string line;
    while (source_file >> line) {
      hdf_filenames_.push_back(line);
    }
  }
  source_file.close();
  num_files_ = hdf_filenames_.size();
  current_file_ = 0;
  LOG(INFO) << "Number of files: " << num_files_;

  // Load the first HDF5 file and initialize the line counter.
  LoadHDF5FileData(hdf_filenames_[current_file_].c_str());
  current_row_ = 0;

  // Reshape blobs.
  const int batch_size = this->layer_param_.hdf5_data_param().batch_size();
  int top_size = this->layer_param_.top_size();
  for (int i = 0; i < top_size; ++i) {
       (*top)[i]->Reshape(batch_size, hdf_blobs_[i]->channels(),
                          hdf_blobs_[i]->height(), hdf_blobs_[i]->width());
       LOG(INFO) << "output '" << this->layer_param_.top(i) << "' size: "
           << (*top)[i]->num() << "," << (*top)[i]->channels()
           << "," << (*top)[i]->height() << ","
           << (*top)[i]->width();
  }
}

template <typename Dtype>
void HDF5DataLayer<Dtype>::Forward_cpu(const vector<Blob<Dtype>*>& bottom,
      vector<Blob<Dtype>*>* top) {
  const int batch_size = this->layer_param_.hdf5_data_param().batch_size();

  for (int i = 0; i < batch_size; ++i, ++current_row_) {
    if (current_row_ == hdf_blobs_[0]->num()) {
      if (num_files_ > 1) {
        current_file_ += 1;
        if (current_file_ == num_files_) {
          current_file_ = 0;
          LOG(INFO) << "looping around to first file";
        }
        LoadHDF5FileData(hdf_filenames_[current_file_].c_str());
      }
      current_row_ = 0;
    }
    for (int j = 0; j < this->layer_param_.top_size(); ++j) {
        int data_count = (*top)[j]->count() / (*top)[j]->num();
        caffe_copy(data_count,
                &hdf_blobs_[j]->cpu_data()[current_row_ * data_count],
                &(*top)[j]->mutable_cpu_data()[i * data_count]);
    }
  }
}

#ifdef CPU_ONLY
STUB_GPU_FORWARD(HDF5DataLayer, Forward);
#endif

INSTANTIATE_CLASS(HDF5DataLayer);

}  // namespace caffe