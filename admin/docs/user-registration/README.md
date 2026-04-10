# User Registration Flow

Tài liệu này tách riêng nghiệp vụ tạo user thành 2 luồng:

- `End User Register`: người chơi tự đăng ký trên màn end user
- `ERP Create User`: quản trị viên tạo user thủ công trên admin ERP

Mục tiêu:

- giữ form end user tối giản
- cho ERP đủ quyền kiểm soát nghiệp vụ và audit
- tránh trộn logic self-service và backoffice vào cùng một mô tả

Tài liệu liên quan:

- [End User Register](./end-user-register.md)
- [ERP Create User](./erp-create-user.md)
- [Bản ghi cần insert](./insert-records.md)
