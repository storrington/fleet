import React from "react";

import { ITeamSummary } from "interfaces/team";
import Button from "components/buttons/Button";
import DataError from "components/DataError";
import Modal from "components/Modal";
import Spinner from "components/Spinner";

import PlatformWrapper from "./PlatformWrapper/PlatformWrapper";
import DownloadInstallers from "./DownloadInstallers/DownloadInstallers";

const baseClass = "add-hosts-modal";

interface IAddHostsModal {
  currentTeam?: ITeamSummary;
  enrollSecret?: string;
  isLoading: boolean;
  isSandboxMode?: boolean;
  onCancel: () => void;
  openEnrollSecretModal?: () => void;
}

const AddHostsModal = ({
  currentTeam,
  enrollSecret,
  isLoading,
  isSandboxMode,
  onCancel,
  openEnrollSecretModal,
}: IAddHostsModal): JSX.Element => {
  const onManageEnrollSecretsClick = () => {
    onCancel();
    openEnrollSecretModal && openEnrollSecretModal();
  };

  const renderModalContent = () => {
    if (isLoading) {
      return <Spinner />;
    }
    if (!enrollSecret) {
      return (
        <DataError>
          <span className="info__data">
            You have no enroll secrets.{" "}
            {openEnrollSecretModal ? (
              <Button onClick={onManageEnrollSecretsClick} variant="text-link">
                Manage enroll secrets
              </Button>
            ) : (
              "Manage enroll secrets"
            )}{" "}
            to enroll hosts to{" "}
            <b>{currentTeam?.id ? currentTeam.name : "Fleet"}</b>.
          </span>
        </DataError>
      );
    }

    // TODO: Currently, prepacked installers in Fleet Sandbox use the global enroll secret,
    // and Fleet Sandbox runs Fleet Free so the currentTeam check here is an
    // additional precaution/reminder to revisit this in connection with future changes.
    // See https://github.com/fleetdm/fleet/issues/4970#issuecomment-1187679407.
    return isSandboxMode && !currentTeam ? (
      <DownloadInstallers onCancel={onCancel} enrollSecret={enrollSecret} />
    ) : (
      <PlatformWrapper onCancel={onCancel} enrollSecret={enrollSecret} />
    );
  };

  return (
    <Modal onExit={onCancel} title={"Add hosts"} className={baseClass}>
      {renderModalContent()}
    </Modal>
  );
};

export default AddHostsModal;