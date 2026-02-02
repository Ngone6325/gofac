package main

import (
	"gofac/di"
	"gofac/model"
)

func main() {
	// 1. 注册3种生命周期服务
	di.MustRegisterAs(model.NewUserRepo, (*model.IUserRepo)(nil), di.Singleton)
	di.MustRegisterAs(model.NewUserLog, (*model.IUserLog)(nil), di.Scoped)
	di.MustRegisterAs(model.NewUserService, (*model.IUserService)(nil), di.Transient)

	// 3. 作用域1：验证Scoped作用域内唯一、Transient每次新建、Singleton全局唯一
	println("\n=== 作用域1 操作 ===")
	scope1 := di.GlobalNewScope()
	// 作用域1第一次获取
	log1_1 := di.ScopeMustGet[model.IUserLog](scope1)
	svc1_1 := di.ScopeMustGet[model.IUserService](scope1)
	repo1_1 := di.ScopeMustGet[model.IUserRepo](scope1)
	println("作用域1-LogUUID：", log1_1.GetLogUUID())
	println("作用域1-SvcUUID：", svc1_1.(*model.UserService).UUID)
	println("作用域1-RepoUUID：", repo1_1.GetRepoUUID())

	// 作用域1第二次获取
	log1_2 := di.ScopeMustGet[model.IUserLog](scope1)
	svc1_2 := di.ScopeMustGet[model.IUserService](scope1)
	repo1_2 := di.ScopeMustGet[model.IUserRepo](scope1)
	println("作用域1-LogUUID（二次获取）：", log1_2.GetLogUUID())              // 相同（Scoped）
	println("作用域1-SvcUUID（二次获取）：", svc1_2.(*model.UserService).UUID) // 不同（Transient）
	println("作用域1-RepoUUID（二次获取）：", repo1_2.GetRepoUUID())           // 相同（Singleton）

	// 4. 作用域2：验证Scoped作用域间隔离
	println("\n=== 作用域2 操作（验证Scoped隔离）===")
	scope2 := di.GlobalNewScope()
	log2_1 := di.ScopeMustGet[model.IUserLog](scope2)
	svc2_1 := di.ScopeMustGet[model.IUserService](scope2)
	repo2_1 := di.ScopeMustGet[model.IUserRepo](scope2)
	println("作用域2-LogUUID：", log2_1.GetLogUUID())              // 不同（Scoped隔离）
	println("作用域2-SvcUUID：", svc2_1.(*model.UserService).UUID) // 全新（Transient）
	println("作用域2-RepoUUID：", repo2_1.GetRepoUUID())           // 相同（Singleton全局唯一）
}
