using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BTCPayServer.Abstractions.Extensions;
using BTCPayServer.Abstractions.Models;
using BTCPayServer.Models;
using BTCPayServer.Models.ServerViewModels;
using BTCPayServer.Storage.Models;
using BTCPayServer.Storage.Services.Providers.AmazonS3Storage;
using BTCPayServer.Storage.Services.Providers.AmazonS3Storage.Configuration;
using BTCPayServer.Storage.Services.Providers.AzureBlobStorage;
using BTCPayServer.Storage.Services.Providers.AzureBlobStorage.Configuration;
using BTCPayServer.Storage.Services.Providers.FileSystemStorage;
using BTCPayServer.Storage.Services.Providers.FileSystemStorage.Configuration;
using BTCPayServer.Storage.Services.Providers.GoogleCloudStorage;
using BTCPayServer.Storage.Services.Providers.GoogleCloudStorage.Configuration;
using BTCPayServer.Storage.Services.Providers.Models;
using BTCPayServer.Storage.ViewModels;
using BTCPayServer.Views;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.Rendering;
using Microsoft.AspNetCore.Mvc.ViewFeatures;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace BTCPayServer.Controllers
{
    public partial class ServerController
    {
        [HttpGet("server/files/{fileId?}")]
        public async Task<IActionResult> Files(string fileId = null, bool multiple = false)
        {
            if (string.IsNullOrEmpty(fileId))
            {
                var model = new ViewFilesViewModel()
                {
                    Files = await _StoredFileRepository.GetFiles(),
                    SelectedFileIds = null,
                    DirectFileUrls = null,
                    StorageConfigured = (await _SettingsRepository.GetSettingAsync<StorageSettings>()) != null
                };
                return View(model);
            }
            else
            {
                List<string> fileIds;
                if (multiple)
                {
                    fileIds = JsonConvert.DeserializeObject<List<string>>(fileId);
                }
                else
                {
                    fileIds = new List<string>();
                    fileIds.Add(fileId);
                }

                List<string> fileUrlList = (fileIds == null || fileIds.Count == 0) ? null : new List<string>();
                bool allFilesExist = true;
                foreach (string filename in fileIds)
                {
                    string fileUrl = await _FileService.GetFileUrl(Request.GetAbsoluteRootUri(), filename);
                    if (fileUrl == null)
                    {
                        allFilesExist = false;
                        break;
                    }
                    fileUrlList.Add(fileUrl);
                }

                if (!allFilesExist)
                {
                    return View(
                        new ViewFilesViewModel()
                        {
                            Files = await _StoredFileRepository.GetFiles(),
                            SelectedFileIds = null,
                            DirectFileUrls = null,
                            StorageConfigured = (await _SettingsRepository.GetSettingAsync<StorageSettings>()) != null
                        }
                    );
                }

                var model = new ViewFilesViewModel()
                {
                    Files = await _StoredFileRepository.GetFiles(),
                    SelectedFileIds = (fileIds == null || fileUrlList == null || fileUrlList.Count != fileIds.Count) ? null : fileIds,
                    DirectFileUrls = (fileIds == null || fileUrlList == null || fileUrlList.Count != fileIds.Count) ? null : fileUrlList,
                    StorageConfigured = (await _SettingsRepository.GetSettingAsync<StorageSettings>()) != null
                };
                return View(model);
            }
        }

        [HttpGet("server/files/{fileId}/delete")]
        public async Task<IActionResult> DeleteFile(string fileId)
        {
            try
            {
                await _FileService.RemoveFile(fileId, null);
                return RedirectToAction(nameof(Files), new
                {
                    fileId = "",
                    statusMessage = "File removed"
                });
            }
            catch (Exception e)
            {
                TempData.SetStatusMessageModel(new StatusMessageModel()
                {
                    Severity = StatusMessageModel.StatusSeverity.Error,
                    Message = e.Message
                });
                return RedirectToAction(nameof(Files));
            }
        }

        [HttpGet("server/files/{fileId}/tmp")]
        public async Task<IActionResult> CreateTemporaryFileUrl(string fileId)
        {
            var file = await _StoredFileRepository.GetFile(fileId);

            if (file == null)
            {
                return NotFound();
            }

            return View(new CreateTemporaryFileUrlViewModel());
        }

        [HttpPost("server/files/{fileId}/tmp")]
        public async Task<IActionResult> CreateTemporaryFileUrl(string fileId,
            CreateTemporaryFileUrlViewModel viewModel)
        {
            if (viewModel.TimeAmount <= 0)
            {
                ModelState.AddModelError(nameof(viewModel.TimeAmount), "Time must be at least 1");
            }

            if (!ModelState.IsValid)
            {
                return View(viewModel);
            }

            var file = await _StoredFileRepository.GetFile(fileId);

            if (file == null)
            {
                return NotFound();
            }

            var expiry = DateTimeOffset.UtcNow;
            switch (viewModel.TimeType)
            {
                case CreateTemporaryFileUrlViewModel.TmpFileTimeType.Seconds:
                    expiry = expiry.AddSeconds(viewModel.TimeAmount);
                    break;
                case CreateTemporaryFileUrlViewModel.TmpFileTimeType.Minutes:
                    expiry = expiry.AddMinutes(viewModel.TimeAmount);
                    break;
                case CreateTemporaryFileUrlViewModel.TmpFileTimeType.Hours:
                    expiry = expiry.AddHours(viewModel.TimeAmount);
                    break;
                case CreateTemporaryFileUrlViewModel.TmpFileTimeType.Days:
                    expiry = expiry.AddDays(viewModel.TimeAmount);
                    break;
                default:
                    throw new ArgumentOutOfRangeException();
            }

            var url = await _FileService.GetTemporaryFileUrl(Request.GetAbsoluteRootUri(), fileId, expiry, viewModel.IsDownload);
            TempData.SetStatusMessageModel(new StatusMessageModel()
            {
                Severity = StatusMessageModel.StatusSeverity.Success,
                Html = $"Generated Temporary Url for file {file.FileName} which expires at {expiry.ToBrowserDate()}. <a href='{url}' target='_blank'>{url}</a>"
            });
            return RedirectToAction(nameof(Files), new
            {
                fileId
            });

        }

        public class CreateTemporaryFileUrlViewModel
        {
            public enum TmpFileTimeType
            {
                Seconds,
                Minutes,
                Hours,
                Days
            }
            public int TimeAmount { get; set; }
            public TmpFileTimeType TimeType { get; set; }
            public bool IsDownload { get; set; }
        }

        
        [HttpPost("server/files/upload")]
        public async Task<IActionResult> CreateFiles(List<IFormFile> files)
        {
            if (files != null && files.Count > 0)
            {
                int invalidFileNameCount = 0;
                List<string> fileIds = new List<string>();
                foreach (IFormFile file in files)
                {
                    if (!file.FileName.IsValidFileName())
                    {
                        invalidFileNameCount++;
                        continue;
                    }
                    var newFile = await _FileService.AddFile(file, GetUserId());
                    fileIds.Add(newFile.Id);
                }

                StatusMessageModel.StatusSeverity statusMessageSeverity;
                string statusMessage; 

                if (invalidFileNameCount == 0)
                {
                    statusMessage = "Files Added Successfully";
                    statusMessageSeverity = StatusMessageModel.StatusSeverity.Success;
                }
                else if (invalidFileNameCount > 0 && invalidFileNameCount < files.Count)
                {
                    statusMessage = $"{files.Count - invalidFileNameCount} files were added. {invalidFileNameCount} files had invalid names";
                    statusMessageSeverity = StatusMessageModel.StatusSeverity.Error;
                }
                else
                {
                    statusMessage = $"Files could not be added due to invalid names";
                    statusMessageSeverity = StatusMessageModel.StatusSeverity.Error;
                }

                this.TempData.SetStatusMessageModel(new StatusMessageModel()
                    {
                        Message = statusMessage,
                        Severity = statusMessageSeverity
                    });

                if (fileIds.Count == 1)
                {
                    return RedirectToAction(nameof(Files), new
                    {
                        statusMessage = "File added!",
                        fileId = fileIds[0]
                    });
                }
                else
                {
                    return RedirectToAction(nameof(Files));
                }
                    multiple = true
                });
            }
            else
            {
                return RedirectToAction(nameof(Files));
            }
        }

        private string GetUserId()
        {
            return _UserManager.GetUserId(ControllerContext.HttpContext.User);
        }

        [HttpGet("server/storage")]
        public async Task<IActionResult> Storage(bool forceChoice = false)
        {
            var savedSettings = await _SettingsRepository.GetSettingAsync<StorageSettings>();
            if (forceChoice || savedSettings == null)
            {
                var providersList = _StorageProviderServices.Select(a =>
                    new SelectListItem(a.StorageProvider().ToString(), a.StorageProvider().ToString())
                );

                return View(new ChooseStorageViewModel()
                {
                    ProvidersList = providersList,
                    ShowChangeWarning = savedSettings != null,
                    Provider = savedSettings?.Provider ?? BTCPayServer.Storage.Models.StorageProvider.FileSystem
                });
            }

            return RedirectToAction(nameof(StorageProvider), new
            {
                provider = savedSettings.Provider
            });
        }

        [HttpPost("server/storage")]
        public IActionResult Storage(StorageSettings viewModel)
        {
            return RedirectToAction("StorageProvider", "Server", new
            {
                provider = viewModel.Provider.ToString()
            });
        }

        [HttpGet("server/storage/{provider}")]
        public async Task<IActionResult> StorageProvider(string provider)
        {
            if (!Enum.TryParse(typeof(StorageProvider), provider, out var storageProvider))
            {
                TempData.SetStatusMessageModel(new StatusMessageModel()
                {
                    Severity = StatusMessageModel.StatusSeverity.Error,
                    Message = $"{provider} provider is not supported"
                });
                return RedirectToAction(nameof(Storage));
            }

            var data = (await _SettingsRepository.GetSettingAsync<StorageSettings>()) ?? new StorageSettings();

            var storageProviderService =
                _StorageProviderServices.SingleOrDefault(service => service.StorageProvider().Equals(storageProvider));

            switch (storageProviderService)
            {
                case null:
                    TempData.SetStatusMessageModel(new StatusMessageModel()
                    {
                        Severity = StatusMessageModel.StatusSeverity.Error,
                        Message = $"{storageProvider} is not supported"
                    });
                    return RedirectToAction(nameof(Storage));
                case AzureBlobStorageFileProviderService fileProviderService:
                    return View(nameof(EditAzureBlobStorageStorageProvider),
                        fileProviderService.GetProviderConfiguration(data));

                case AmazonS3FileProviderService fileProviderService:
                    return View(nameof(EditAmazonS3StorageProvider),
                        fileProviderService.GetProviderConfiguration(data));
                case GoogleCloudStorageFileProviderService fileProviderService:
                    return View(nameof(EditGoogleCloudStorageStorageProvider),
                        fileProviderService.GetProviderConfiguration(data));
                case FileSystemFileProviderService fileProviderService:
                    if (data.Provider != BTCPayServer.Storage.Models.StorageProvider.FileSystem)
                    {
                        _ = await SaveStorageProvider(new FileSystemStorageConfiguration(),
                            BTCPayServer.Storage.Models.StorageProvider.FileSystem);
                    }

                    return View(nameof(EditFileSystemStorageProvider),
                        fileProviderService.GetProviderConfiguration(data));
            }

            return NotFound();
        }


        [HttpPost("server/storage/AzureBlobStorage")]
        public async Task<IActionResult> EditAzureBlobStorageStorageProvider(AzureBlobStorageConfiguration viewModel)
        {
            return await SaveStorageProvider(viewModel, BTCPayServer.Storage.Models.StorageProvider.AzureBlobStorage);
        }

        [HttpPost("server/storage/AmazonS3")]
        public async Task<IActionResult> EditAmazonS3StorageProvider(AmazonS3StorageConfiguration viewModel)
        {
            return await SaveStorageProvider(viewModel, BTCPayServer.Storage.Models.StorageProvider.AmazonS3);
        }

        [HttpPost("server/storage/GoogleCloudStorage")]
        public async Task<IActionResult> EditGoogleCloudStorageStorageProvider(
            GoogleCloudStorageConfiguration viewModel)
        {
            return await SaveStorageProvider(viewModel, BTCPayServer.Storage.Models.StorageProvider.GoogleCloudStorage);
        }

        [HttpPost("server/storage/FileSystem")]
        public async Task<IActionResult> EditFileSystemStorageProvider(FileSystemStorageConfiguration viewModel)
        {
            return await SaveStorageProvider(viewModel, BTCPayServer.Storage.Models.StorageProvider.FileSystem);
        }

        private async Task<IActionResult> SaveStorageProvider(IBaseStorageConfiguration viewModel,
            StorageProvider storageProvider)
        {
            if (!ModelState.IsValid)
            {
                return View(viewModel);
            }

            var data = (await _SettingsRepository.GetSettingAsync<StorageSettings>()) ?? new StorageSettings();
            data.Provider = storageProvider;
            data.Configuration = JObject.FromObject(viewModel);
            await _SettingsRepository.UpdateSetting(data);
            TempData.SetStatusMessageModel(new StatusMessageModel()
            {
                Severity = StatusMessageModel.StatusSeverity.Success,
                Message = "Storage settings updated successfully"
            });
            return View(viewModel);
        }
    }
}