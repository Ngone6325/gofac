package model

import "fmt"

// IUserRepo 用户仓库接口（Singleton）
type IUserRepo interface {
	GetUserID() int64
	GetRepoUUID() string // 获取实例自身的唯一地址
}

// UserRepo IUserRepo实现（Singleton）
type UserRepo struct {
	DBConfig string
	UUID     string // 存储自身实例的指针地址
}

// NewUserRepo 构造函数：初始化时记录自身实例的指针地址
func NewUserRepo() *UserRepo {
	fmt.Println("[构造函数] UserRepo (Singleton) 执行")
	repo := &UserRepo{
		DBConfig: "mysql:127.0.0.1:3306/gofac?charset=utf8",
	}
	// 核心修复：取当前实例自身的指针地址（%p取&repo的地址）
	repo.UUID = fmt.Sprintf("%p", repo)
	return repo
}

func (r *UserRepo) GetUserID() int64    { return 10086 }
func (r *UserRepo) GetRepoUUID() string { return r.UUID }

// IUserService 用户服务接口（Transient）
type IUserService interface {
	GetUserName() string
	GetRepoUUID() string
}

// UserService IUserService实现（Transient）
type UserService struct {
	Repo IUserRepo
	UUID string // 存储自身实例的指针地址
}

// NewUserService 构造函数：记录自身实例的指针地址
func NewUserService(repo IUserRepo) *UserService {
	fmt.Println("[构造函数] UserService (Transient) 执行")
	svc := &UserService{
		Repo: repo,
	}
	svc.UUID = fmt.Sprintf("%p", svc)
	return svc
}

func (s *UserService) GetUserName() string { return fmt.Sprintf("user_%d", s.Repo.GetUserID()) }
func (s *UserService) GetRepoUUID() string { return s.Repo.GetRepoUUID() }

// IUserLog 用户日志接口（Scoped）
type IUserLog interface {
	LogUserID() string
	GetLogUUID() string // 获取自身实例的唯一地址
}

// UserLog IUserLog实现（Scoped）
type UserLog struct {
	Repo IUserRepo
	UUID string // 存储自身实例的指针地址
}

// NewUserLog 构造函数：记录自身实例的指针地址
func NewUserLog(repo IUserRepo) *UserLog {
	fmt.Println("[构造函数] UserLog (Scoped) 执行")
	log := &UserLog{
		Repo: repo,
	}
	log.UUID = fmt.Sprintf("%p", log)
	return log
}

func (l *UserLog) LogUserID() string  { return fmt.Sprintf("user_log: user_id=%d", l.Repo.GetUserID()) }
func (l *UserLog) GetLogUUID() string { return l.UUID }
