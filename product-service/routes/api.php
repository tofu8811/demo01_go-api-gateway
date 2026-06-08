<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\ProductController;
use App\Http\Controllers\ServiceAccountController;

Route::middleware('gateway.secret')->group(function () {
    Route::post('/product/create', [ProductController::class, 'create']);
    Route::put('/product/update/{id}', [ProductController::class, 'update']);
    Route::delete('/product/delete/{id}', [ProductController::class, 'delete']);
});
// tạm bỏ qua phần auth
Route::get('/products', [ProductController::class, 'index']);
Route::get('/product/{id}', [ProductController::class, 'show']);


Route::post('service-accounts/token', [ServiceAccountController::class, 'issueToken']);
