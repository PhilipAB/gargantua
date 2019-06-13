package vmsetcontroller

import (
"fmt"
"github.com/golang/glog"
hfv1 "github.com/hobbyfarm/gargantua/pkg/apis/hobbyfarm.io/v1"
hfClientset "github.com/hobbyfarm/gargantua/pkg/client/clientset/versioned"
hfInformers "github.com/hobbyfarm/gargantua/pkg/client/informers/externalversions"
hfListers "github.com/hobbyfarm/gargantua/pkg/client/listers/hobbyfarm.io/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"github.com/hobbyfarm/gargantua/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"math/rand"
	"strings"
	"time"
)

type VirtualMachineSetController struct {
	hfClientSet *hfClientset.Clientset
	vmSetWorkqueue workqueue.RateLimitingInterface
	vmWorkqueue workqueue.RateLimitingInterface

	vmSetLister hfListers.VirtualMachineSetLister
	vmLister hfListers.VirtualMachineLister
	envLister hfListers.EnvironmentLister
	vmTemplateLister hfListers.VirtualMachineTemplateLister

	vmSetSynced cache.InformerSynced
	vmSynced cache.InformerSynced
	envSynced cache.InformerSynced
	vmTemplateSynced cache.InformerSynced
}
const (
	vmEnvironmentIndex = "vm.vmclaim.controllers.hobbyfarm.io/environment-index"
)

func NewVirtualMachineSetController(hfClientSet *hfClientset.Clientset, hfInformerFactory hfInformers.SharedInformerFactory) (*VirtualMachineSetController, error) {
	vmSetController := VirtualMachineSetController{}
	vmSetController.hfClientSet = hfClientSet
	vmSetController.vmSetSynced = hfInformerFactory.Hobbyfarm().V1().VirtualMachineSets().Informer().HasSynced
	vmSetController.vmSynced = hfInformerFactory.Hobbyfarm().V1().VirtualMachines().Informer().HasSynced
	vmSetController.envSynced = hfInformerFactory.Hobbyfarm().V1().Environments().Informer().HasSynced
	vmSetController.vmTemplateSynced = hfInformerFactory.Hobbyfarm().V1().VirtualMachineTemplates().Informer().HasSynced

	vmSetController.vmSetWorkqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "VMSet")
	vmSetController.vmWorkqueue = workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "VM")
	//vmClaimController.hfInformerFactory = hfInformerFactory

	vmSetController.vmSetLister = hfInformerFactory.Hobbyfarm().V1().VirtualMachineSets().Lister()
	vmSetController.vmLister = hfInformerFactory.Hobbyfarm().V1().VirtualMachines().Lister()
	vmSetController.envLister = hfInformerFactory.Hobbyfarm().V1().Environments().Lister()
	vmSetController.vmTemplateLister = hfInformerFactory.Hobbyfarm().V1().VirtualMachineTemplates().Lister()

	vmSetInformer := hfInformerFactory.Hobbyfarm().V1().VirtualMachineSets().Informer()

	vmSetInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: vmSetController.enqueueVMSet,
		UpdateFunc: func(old, new interface{}) {
			vmSetController.enqueueVMSet(new)
		},
		DeleteFunc: vmSetController.enqueueVMSet,
	}, time.Second * 30)

	vmInformer := hfInformerFactory.Hobbyfarm().V1().VirtualMachines().Informer()

	vmInformer.AddEventHandlerWithResyncPeriod(cache.ResourceEventHandlerFuncs{
		AddFunc: vmSetController.handleVM,
		UpdateFunc: func(old, new interface{}) {
			vmSetController.handleVM(new)
		},
		DeleteFunc: vmSetController.handleVM,
	},  time.Second * 30)
	return &vmSetController, nil
}

func (v *VirtualMachineSetController) handleVM(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			glog.Errorf("error decoding object, invalid type")
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			glog.Errorf("error decoding object tombstone, invalid type")
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a VirtualMachineSet, we should not do anything more
		// with it.
		if ownerRef.Kind != "VirtualMachineSet" {
			return
		}

		vms, err := v.vmSetLister.Get(ownerRef.Name)
		if err != nil {
			klog.V(4).Infof("ignoring orphaned object '%s' of vmset '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		v.enqueueVMSet(vms)
		return
	}
}

func (v *VirtualMachineSetController) enqueueVMSet(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		//utilruntime.HandleError(err)
		return
	}
	glog.V(8).Infof("Enqueueing vm set %s", key)
	v.vmSetWorkqueue.AddRateLimited(key)
}

func (v *VirtualMachineSetController) Run(stopCh <-chan struct{}) error {
	defer v.vmSetWorkqueue.ShutDown()

	glog.V(4).Infof("Starting vm set")
	glog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, v.vmSynced, v.vmSetSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}
	glog.Info("Starting vm set controller worker")
	go wait.Until(v.runVMSetWorker, time.Second, stopCh)
	glog.Info("Started vm set controller worker")
	//if ok := cache.WaitForCacheSync(stopCh, )
	<-stopCh
	return nil
}

func (v *VirtualMachineSetController) runVMSetWorker() {
	for v.processNextVMSet() {

	}
}

func (v *VirtualMachineSetController) processNextVMSet() bool {
	obj, shutdown := v.vmSetWorkqueue.Get()

	glog.V(4).Infof("processing VMSet")

	if shutdown {
		return false
	}
	err := func() error {
		defer v.vmSetWorkqueue.Done(obj)
		glog.V(4).Infof("processing vm in env controller: %v", obj)
		_, objName, err := cache.SplitMetaNamespaceKey(obj.(string)) // this is actually not necessary because VM's are not namespaced yet...
		if err != nil {
			glog.Errorf("error while splitting meta namespace key %v", err)
			//e.vmWorkqueue.AddRateLimited(obj)
			return nil
		}

		vmSet, err := v.vmSetLister.Get(objName)
		if err != nil {
			glog.Errorf("error while retrieving virtual machine set %s, likely deleted %v", objName, err)
			v.vmSetWorkqueue.Forget(obj)
			return nil
		}

		err = v.reconcileVirtualMachineSet(vmSet)
		if err != nil {
			glog.Error(err)
		}
		v.vmSetWorkqueue.Forget(obj)
		glog.V(4).Infof("vm set processed by vmset controller %v", objName)

		return nil

	}()

	if err != nil {
		return true
	}

	return true
}

func (v *VirtualMachineSetController) reconcileVirtualMachineSet(vmset *hfv1.VirtualMachineSet) error {

	if vmset.Spec.Count >= vmset.Status.ProvisionedCount { // if desired count is greater than the current provisioned
		// 1. let's check the environment to see if there is available capacity
		// 2. if available capacity is available let's create new VM's
		env, err := v.envLister.Get(vmset.Spec.Environment)
		if err != nil {
			if apierrors.IsNotFound(err) {
				glog.Errorf("environment invalid")
			}
			return err
		}

		vmt, err := v.vmTemplateLister.Get(vmset.Spec.VMTemplate)

		availableStorage := env.Status.Capacity.Storage - env.Status.Used.Storage
		availableMemory := env.Status.Capacity.Memory - env.Status.Used.Memory
		availableCPU := env.Status.Capacity.CPU - env.Status.Used.CPU
		glog.V(5).Infof("available cpu: %d, mem: %d, storage: %d", availableCPU, availableMemory, availableStorage)
		numberOfVMCPU := availableCPU/vmt.Spec.Resources.CPU
		numberofVMMemory := availableMemory/vmt.Spec.Resources.Memory
		numberofVMStorage := availableStorage/vmt.Spec.Resources.Storage

		glog.V(5).Infof("calculated cpu: %d, mem: %d, storage: %d", numberOfVMCPU, numberofVMMemory, numberofVMStorage)

		capable := numberOfVMCPU
		if numberofVMMemory < capable {
			capable = numberofVMMemory
		}
		if numberofVMStorage < capable {
			capable = numberofVMStorage
		}

		glog.V(5).Infof("counted number of capable vms for env %s is: %d", env.Name, capable)

		needed := vmset.Spec.Count - vmset.Status.ProvisionedCount


		glog.V(5).Infof("counted number of needed vms for env %s is: %d", env.Name, needed)

		pCount := 0
		if needed > capable {
			pCount = capable
		} else {
			pCount = needed
		}

		glog.V(5).Infof("provisioning %d vms", pCount)
		// this code is so... verbose...
		for i := 0; i < pCount; i++ {
			vmName := strings.Join([]string{vmset.Spec.BaseName, fmt.Sprintf("%08x", rand.Uint32())}, "-")
			vm := &hfv1.VirtualMachine{
				ObjectMeta: metav1.ObjectMeta{
					Name: vmName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "v1",
							Kind: "VirtualMachineSet",
							Name: vmset.Name,
							UID: vmset.UID,
						},
						{
							APIVersion: "v1",
							Kind: "Environment",
							Name: env.Name,
							UID: env.UID,
						},
					},
					Labels: map[string]string{
						"vmset": vmset.Name,
					},
				},
				Spec: hfv1.VirtualMachineSpec{
					Id: vmName,
					VirtualMachineTemplateId: vmt.Spec.Id,
					KeyPair: "",
					VirtualMachineClaimId: "",
					UserId: "",
					Provision: true,
					VirtualMachineSetId: vmset.Name,
				},
				Status: hfv1.VirtualMachineStatus{
					Status: hfv1.VmStatusRFP,
					Allocated: false,
					Tainted: false,
					PublicIP: "",
					PrivateIP: "",
					EnvironmentId: env.Name,
					Hostname: "",
				},
			}
			vm, err := v.hfClientSet.HobbyfarmV1().VirtualMachines().Create(vm)
			if err != nil {
				glog.Error(err)
			}

			err = v.verifyVM(vm)
			if err != nil {
				glog.Error(err)
			}
		}
	}

	// no matter what we should list the vm's and delete the ones that are ready for deletion

	vms, err := v.vmLister.List(labels.Set{
		"vmset":  string(vmset.Name),
	}.AsSelector())

	if err != nil {
		glog.Errorf("error while retrieving vms owned by vmset %s", vmset.Name)
	}

	for _, x := range vms {
		if x.DeletionTimestamp == nil && x.Status.Tainted {
			err := v.deleteVM(x)
			if err != nil {
				glog.Error(err)
			}
		}
	}

	vms, err = v.vmLister.List(labels.Set{
		"vmset":  string(vmset.Name),
	}.AsSelector())

	if err != nil {
		glog.Errorf("error while retrieving vms owned by vmset %s", vmset.Name)
	}

	provisionedCount := 0
	activeCount := 0
	for _, x := range vms {
		if x.DeletionTimestamp == nil {
			activeCount++
		}
		provisionedCount++
	}

	err = v.updateVMSetCount(vmset.Name, activeCount, provisionedCount)

	return nil
}

func (v *VirtualMachineSetController) deleteVM(vm *hfv1.VirtualMachine) error {
	err := v.hfClientSet.HobbyfarmV1().VirtualMachines().Delete(vm.Name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	for i := 0; i < 25; i++ {
		vmFromLister, err := v.vmLister.Get(vm.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if vmFromLister.DeletionTimestamp != nil {
			return nil // we are done waiting for delete to happen finalizers can happen in the background doesn't matter
		}
	}
	return nil
}


func (v *VirtualMachineSetController) createVM(vm *hfv1.VirtualMachine) error {
	vm, err := v.hfClientSet.HobbyfarmV1().VirtualMachines().Create(vm)
	if err != nil {
		return err
	}
	for i := 0; i < 25; i++ {
		vmFromLister, err := v.vmLister.Get(vm.Name)
		if err != nil {
			if apierrors.IsNotFound(err) {
				glog.V(5).Infof("vm not in lister yet %s", vm.Name)
				break
			}
		}
		if util.ResourceVersionAtLeast(vmFromLister.ResourceVersion, vm.ResourceVersion) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (v *VirtualMachineSetController) updateVMSetCount(vmSetName string, active int, prov int) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		result, getErr := v.hfClientSet.HobbyfarmV1().VirtualMachineSets().Get(vmSetName, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Error retrieving latest version of Virtual Machine Set %s: %v", vmSetName, getErr)
		}

		result.Status.ProvisionedCount = prov
		result.Status.AvailableCount = active

		vms, updateErr := v.hfClientSet.HobbyfarmV1().VirtualMachineSets().Update(result)
		if updateErr != nil {
			glog.Error(updateErr)
			return updateErr
		}

		err := v.verifyVMSet(vms)
		if err != nil {
			glog.Error(err)
		}
		glog.V(4).Infof("updated result for virtual machine set")
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Error updating Virtual Machine Set: %s, %v", vmSetName, retryErr)
	}

	return nil
}

func (v *VirtualMachineSetController) verifyVMSet(vms *hfv1.VirtualMachineSet) error {
	var err error
	glog.V(5).Infof("Verifying vms %s", vms.Name)
	for i := 0; i < 50; i++ {
		var fromCache *hfv1.VirtualMachineSet
		fromCache, err = v.vmSetLister.Get(vms.Name)
		if err != nil {
			glog.Error(err)
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		if util.ResourceVersionAtLeast(fromCache.ResourceVersion, vms.ResourceVersion) {
			glog.V(5).Infof("resource version matched for %s", vms.Name)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	glog.Errorf("resource version didn't match for in time %s", vms.Name)
	return nil

}

func (v *VirtualMachineSetController) verifyVM(vm *hfv1.VirtualMachine) error {
	var err error
	glog.V(5).Infof("Verifying vm %s", vm.Name)
	for i := 0; i < 150000; i++ {
		var fromCache *hfv1.VirtualMachine
		fromCache, err = v.vmLister.Get(vm.Name)
		if err != nil {
			glog.Error(err)
			if apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		if util.ResourceVersionAtLeast(fromCache.ResourceVersion, vm.ResourceVersion) {
			glog.V(5).Infof("resource version matched for %s", vm.Name)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	glog.Errorf("resource version didn't match for in time%s", vm.Name)
	return nil

}