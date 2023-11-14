import nibabel
import io
import PIL.Image
import torch
import random
def zlzheimer_diagnostic_system(is_demo=False):
    from pyecharts.charts import Bar

    nii_path = "./demodata/AD/ADNI_033_S_1308_MR_MP-RAGE__br_raw_20080228150456528_3_S46343_I93697.nii"

    img = nibabel.load(nii_path)
    img = img.get_fdata()
    print(img.shape)
        # (166, 256, 256, 1)
    torch.no_grad()
    test_model = torch.load("./myModel_109.pth", map_location=torch.device('cpu'))
    test_model.eval()

    processed_img = torch.from_numpy(img)
    processed_img = processed_img.squeeze()
    processed_img = processed_img.reshape(1, -1, 256, 256)
    processed_img = processed_img[:, 0:160, :, :].float()
    processed_img = processed_img.reshape((1, 1, -1, 256, 256))

    output = None
    with torch.no_grad():
        output = test_model(processed_img)
    ans_y = output.squeeze().tolist()
    print(ans_y)
    del test_model,processed_img

    from datasets import LABEL_LIST
    if min(ans_y) < 0:
        m = min(ans_y)
        for i in range(len(ans_y)):
            ans_y[i] -= 1.2 * m

    ans = LABEL_LIST[output.argmax(1).item()]
    if ans == 'AD':
        ans += '（阿尔茨海默病）'
    elif ans == 'CN':
        ans += '（认知正常）'
    elif ans == 'MCI':
        ans += '（轻度认知障碍）'
    elif ans == 'EMCI':
        ans += '（早期轻度认知障碍）'
    elif ans == 'LMCI':
        ans += '（晚期轻度认知障碍）'
    #show_result = [pywebio.output.put_markdown("诊断为：\n # " + ans),                       pywebio.output.put_warning('结果仅供参考')]
    print("诊断为："+ans+".结果仅供参考")
if __name__ == "__main__":
   zlzheimer_diagnostic_system() 
